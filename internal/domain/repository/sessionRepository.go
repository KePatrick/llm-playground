package repository

import (
	"context"
	"kepatrick/llm-playground/internal/domain/entity"
)

type SessionRepository interface {
	AppendMessage(ctx context.Context, sessionID string, msg entity.Message) error
	FetchPrevMessage(ctx context.Context, sessionID string) ([]entity.Message, error)
	ExistKey(ctx context.Context, sessionID string) bool
}
