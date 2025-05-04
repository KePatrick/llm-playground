package redis

import (
	"context"
	"encoding/json"
	"kepatrick/llm-playground/internal/domain/entity"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRepo struct {
	Client *redis.Client
}

func NewRedisSessionRepo(client *redis.Client) *RedisSessionRepo {
	return &RedisSessionRepo{Client: client}
}

func (r *RedisSessionRepo) AppendMessage(ctx context.Context, sessionID string, msg entity.Message) error {
	data, _ := json.Marshal(msg)
	return r.Client.RPush(ctx, sessionID, data).Err()
}

func (r *RedisSessionRepo) FetchPrevMessage(ctx context.Context, sessionID string) ([]entity.Message, error) {

	msgJSONs, err := r.Client.LRange(ctx, sessionID, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	// Deserialize JSON into a message structure
	var messages []entity.Message
	for _, msgJSON := range msgJSONs {
		var msg entity.Message
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *RedisSessionRepo) ExistKey(ctx context.Context, sessionID string) bool {
	count, err := r.Client.Exists(ctx, sessionID).Result()
	if err != nil {
		return false
	}
	return count > 0
}
