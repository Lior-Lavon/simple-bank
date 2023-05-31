package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/rs/zerolog/log"
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
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(reportError), // define error callback function
			Logger:       NewLogger(),                         // define custome logger struct for the logger interface
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

func reportError(ctx context.Context, task *asynq.Task, err error) {
	// retried, _ := asynq.GetRetryCount(ctx)
	// maxRetry, _ := asynq.GetMaxRetry(ctx)
	// if retried >= maxRetry {
	// 	err = fmt.Errorf("retry exhausted for task %s: %w", task.Type, err)
	// }
	// errorReportingService.Notify(err)

	log.Error().Err(err).Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("process task failed")
}
