package entity

type Message struct {
	Role       string                   `json:"role"`
	Content    string                   `json:"content"`
	ToolCallID string                   `json:"tool_call_id"`
	ToolCalls  []map[string]interface{} `json:"tool_calls"`
	Timestamp  string
}
