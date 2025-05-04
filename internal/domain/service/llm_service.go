package service

import (
	"context"
	"kepatrick/llm-playground/internal/domain/entity"
)

type LLMResult struct {
	LlmRes        string
	IsToolCall    bool
	ToolCallDepth int
	ReqToken      int
	ResToken      int
	Messages      []entity.Message
}

type LLMService interface {
	StreamingCall(ctx context.Context, messages []entity.Message, writer StreamWriter, lastRslt LLMResult) (LLMResult, error)
}

type StreamWriter interface {
	Write(data string) error
	Done() error
}
