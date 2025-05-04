package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kepatrick/llm-playground/internal/config"
	"kepatrick/llm-playground/internal/domain/entity"
	"kepatrick/llm-playground/internal/domain/service"
	"net/http"
	"os/exec"
	"strings" // String manipulation
	"time"    // Time utilities

	"github.com/pkg/errors"
)

// maxToolCallDepth defines the maximum recursion depth for tool calls
const maxToolCallDepth = 5

// FunctionCall represents a function call with ID, name, and arguments
type FunctionCall struct {
	ID        string          // Unique identifier for the function call
	Name      string          // Name of the function
	Arguments strings.Builder // Arguments for the function call
}

// OpenAILLMService handles interactions with the OpenAI API
type OpenAILLMService struct {
	apiKey string       // API key for authentication
	apiUrl string       // API endpoint URL
	model  string       // model
	client *http.Client // HTTP client for making requests
	tools  []config.Tool
}

// NewOpenAILLMService creates a new instance of OpenAILLMService
func NewOpenAILLMService(key, url, model string, cli *http.Client, tools []config.Tool) *OpenAILLMService {
	return &OpenAILLMService{key, url, model, cli, tools}
}

func (s *OpenAILLMService) StreamingCall(ctx context.Context, messages []entity.Message, writer service.StreamWriter, lastRslt service.LLMResult) (service.LLMResult, error) {
	var builder strings.Builder
	var curReqToken int
	var curResToken int

	depth := lastRslt.ToolCallDepth + 1
	reqTokens := lastRslt.ReqToken
	resTokens := lastRslt.ResToken

	// needFunctionCall := false
	// functionCalls slice
	functionCalls := []*FunctionCall{}

	if depth > maxToolCallDepth {
		return buildLLMRslt("tool calling out of limit", false, depth, reqTokens, resTokens, nil), fmt.Errorf("tool call depth exceeded")
	}
	// Build messages from session and prompt
	msgs, err := s.buildMessages(messages)
	if err != nil {
		return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), fmt.Errorf("build message failed")
	}

	tools := s.prepareReqTools(s.tools)

	// Prepare request body with streaming enabled
	body := map[string]interface{}{
		"model":    s.model,
		"messages": msgs,
		"stream":   true,
		"tools":    tools,
	}

	// if (strings.Contains(s.apiUrl, "openai")) {
	// 	body = map[string]interface{}{
	// 	"model":    s.model,
	// 	"stream_options": map[string]bool{"include_usage": true},
	// 	"messages": msgs,
	// 	"stream":   true,
	// 	"tools":    tools,
	// 	}
	// }
	data, err := json.Marshal(body)

	if err != nil {
		return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), errors.Wrap(err, "marshal stream body")
	}
	// jsonStr, _ := json.MarshalIndent(body, "", "  ")
	// fmt.Printf("reqBody: %s\n", jsonStr)

	// Create HTTP request
	req, _ := http.NewRequestWithContext(ctx, "POST", s.apiUrl, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	res, err := s.client.Do(req)
	if err != nil {
		return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), err
	}
	defer res.Body.Close()

	// Check http status
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), fmt.Errorf("stream upstream error %d: %s", res.StatusCode, string(b))
	}

	// Read response stream
	rd := bufio.NewReader(res.Body)
	for {
		// Read line from stream
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), err
		}

		// Skip non-data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Process chunk
		chunk := strings.TrimPrefix(strings.TrimSpace(line), "data: ")
		if chunk == "[DONE]" {
			if len(functionCalls) > 0 {
				for _, fc := range functionCalls {
					var resStr string
					for _, tool := range s.tools {
						if tool.Function.Name == fc.Name {
							resStr, err = callTool(tool, fc)
							if err != nil {
								resStr = "fail to call tool"
							}
						}
					}

					if err != nil {
						return buildLLMRslt("", false, depth, reqTokens, resTokens, nil), err
					}
					// Add toolCall message to messages slice
					messages = append(messages, entity.Message{
						Role:       "assistant",
						Content:    "",    // No content if functionCall
						ToolCallID: fc.ID, // function call id
						ToolCalls: []map[string]interface{}{{
							"id":   fc.ID,
							"type": "function",
							"function": map[string]interface{}{
								"name":      fc.Name,
								"arguments": fc.Arguments.String(),
							},
						}},
						Timestamp: nowMilli(),
					})
					// Add toolCall result to messages slice
					messages = append(messages, entity.Message{
						Role:       "tool",
						Content:    resStr,
						ToolCallID: fc.ID,
						ToolCalls:  nil,
						Timestamp:  nowMilli(),
					})

				}
				// return with toolcall
				return buildLLMRslt("", true, depth, reqTokens, resTokens, messages), err

			}
			writer.Done()
			break
		}

		// Parse event
		var event map[string]interface{}
		json.Unmarshal([]byte(chunk), &event)

		if usage, ok := event["usage"].(map[string]interface{}); ok && usage != nil {
			if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
				curReqToken = int(promptTokens)
			}
			if completionTokens, ok := usage["completion_tokens"].(float64); ok {
				curResToken = int(completionTokens)
			}
		}

		// Handle tool call
		_, functionCalls = parseToolCall(event, functionCalls)

		// Write content to stream
		content := extractContent(event)
		if content != "" {
			builder.WriteString(content)
			writer.Write(content)
			// Append generate result when \n\n
			if strings.HasSuffix(content, "\n\n") {
				messages = append(messages, entity.Message{
					Role:       "assistant",
					Content:    builder.String(),
					ToolCallID: "",
					ToolCalls:  nil,
					Timestamp:  nowMilli(),
				})
				builder.Reset()
			}
		}
	}
	reqTokens += curReqToken
	resTokens += curResToken
	depth++

	// Flush if builder has remain message
	if rem := builder.String(); rem != "" {
		writer.Write(rem)
		messages = append(messages, entity.Message{
			Role:       "assistant",
			Content:    rem,
			ToolCallID: "",
			ToolCalls:  nil,
			Timestamp:  nowMilli(),
		})
	}

	return buildLLMRslt(builder.String(), false, depth, reqTokens, resTokens, messages), err
}

// buildMessages constructs the message array for API requests
func (s *OpenAILLMService) buildMessages(raw []entity.Message) ([]map[string]interface{}, error) {

	var msgs []map[string]interface{}
	for _, m := range raw {
		entry := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if m.ToolCallID != "" {
			entry["tool_call_id"] = m.ToolCallID
		}
		if m.ToolCalls != nil {
			entry["tool_calls"] = m.ToolCalls
		}
		msgs = append(msgs, entry)
	}
	// fmt.Printf("prevMsg:%v\n", msgs)
	return msgs, nil
}

// Clear scripts for api request
func (s *OpenAILLMService) prepareReqTools(tools []config.Tool) []config.Tool {

	// make return slice
	results := make([]config.Tool, len(tools))
	copy(results, tools)
	for i := range results {
		results[i].Script = ""
	}
	return results
}

// parseToolCall checks if the event contains a tool call
func parseToolCall(evt map[string]interface{}, fcs []*FunctionCall) (bool, []*FunctionCall) {
	chs, ok := evt["choices"].([]interface{})
	if !ok || len(chs) == 0 {
		return false, nil
	}
	delta, ok := chs[0].(map[string]interface{})["delta"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	if tcs, ok := delta["tool_calls"].([]interface{}); ok && len(tcs) > 0 {
		raw := tcs[0].(map[string]interface{})
		fn := raw["function"].(map[string]interface{})
		id, _ := raw["id"].(string)
		name, _ := fn["name"].(string)
		args, _ := fn["arguments"].(string)
		if name != "" {
			fcs = append(fcs, &FunctionCall{ID: id, Name: name})
		} else {
			fcs[len(fcs)-1].Arguments.WriteString(args)
		}
	}
	return false, fcs
}

// extractContent from delta.content
func extractContent(evt map[string]interface{}) string {
	chs, ok := evt["choices"].([]interface{})
	if !ok || len(chs) == 0 {
		return ""
	}
	delta, ok := chs[0].(map[string]interface{})["delta"].(map[string]interface{})
	if !ok {
		return ""
	}
	content, _ := delta["content"].(string)
	// somehow \n just cant work on javascript
	content = strings.ReplaceAll(content, "\n", "[NEWLINE]")
	return content
}

// nowMilli returns the current time in milliseconds as a string
func nowMilli() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

func callTool(tool config.Tool, fc *FunctionCall) (string, error) {
	script := "./scripts/" + tool.Script

	fmt.Printf("tool: %s", tool.Function.Name)
	fmt.Printf("\nscrpit:%s", tool.Script)

	var args map[string]string
	err := json.Unmarshal([]byte(fc.Arguments.String()), &args)
	if err != nil {
		return "", fmt.Errorf("fail to parse arguments to map: %w", err)
	}

	var cmdArgs []string
	for k, v := range args {
		cmdArgs = append(cmdArgs, "--"+k, v)
	}

	cmd := exec.Command(script, cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

func buildLLMRslt(res string, isToolCall bool, depth int, reqTokens int, resToken int, messages []entity.Message) service.LLMResult {
	return service.LLMResult{
		LlmRes:        res,
		IsToolCall:    isToolCall,
		ToolCallDepth: depth,
		ReqToken:      reqTokens,
		ResToken:      resToken,
		Messages:      messages,
	}
}
