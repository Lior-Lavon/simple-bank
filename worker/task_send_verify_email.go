package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

// this task will contain all data of the task that we want to store in Redis
// and later the worker would be able to retreive it from the Queue
type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// Sender : create a new Task and send it to a Redis Queue
func (distributor *RedisTaskDistributed) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {

	// lets serielize the payload into json
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshel task payload : %w", err)
	}

	// create a new Task
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	// send the task to a redis queue
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", taskInfo.Queue).
		Int("max_retry", taskInfo.MaxRetry).
		Msg("enqueued task")

	return nil
}

// Handler: Task handler that receive a task from the asynq, extract the user name, retreive the user and and email
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {

	var payload PayloadSendVerifyEmail

	// get the payload from the task and Unmarshal it to object
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshel payload: %w", asynq.SkipRetry) // SkipRetry: dont retry again
	}

	// retreive the user info from the database using the username
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist: %w", asynq.SkipRetry) // // SkipRetry: dont retry again
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// ToDo : Send the email to the user here
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("processed task!")

	return nil
}
