package ingestion

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type QueueRepository interface {
	PushToQueue(ctx context.Context, queueName string, data []byte) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRepository(client *redis.Client) QueueRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) PushToQueue(ctx context.Context, queueName string, data []byte) error {
	return r.client.RPush(ctx, queueName, data).Err()
}
