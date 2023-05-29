package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

// this file will create the tasks and distribute them to the redis queue
// using interface and struct, allow to mock the functionality for unit-testing
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error
}

// RedisTaskDistributed will implement the interface TaskDistributer
type RedisTaskDistributed struct {
	client *asynq.Client
}

func NewRedisTaskDistributer(redisOpt asynq.RedisConnOpt) TaskDistributor {
	// create new client
	client := asynq.NewClient(redisOpt)

	return &RedisTaskDistributed{
		client: client,
	}
}
