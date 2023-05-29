package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/liorlavon/simplebank/db/sqlc"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// this processor will pickup the task from the Redis queue and process it
type iTaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor will implement the iTaskProcessor interface
type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store // allow the processor to access the db
}

func NewRedisTaskProcessor(redisOpt asynq.RedisConnOpt, store db.Store) iTaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			QueueCritical: 10,
			QueueDefault:  5,
		},
	})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

// register the task with the processorFunc in asynq server
func (processor *RedisTaskProcessor) Start() error {

	mux := asynq.NewServeMux()

	// registe the task and the handler function
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	// start the worker server
	return processor.server.Start(mux)
}
