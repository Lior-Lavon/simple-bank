package gapi

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/validation"
	"github.com/liorlavon/simplebank/worker"

	//"github.com/liorlavon/simplebank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	// validate input
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// hash the password
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to Hash password %s", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hashedPassword,
			Firstname:      req.GetFirstname(),
			Lastname:       req.GetLastname(),
			Email:          req.GetEmail(),
		},
		// implement this function after the create user success
		AfterCreate: func(user db.User) error {
			// send a verification email to the user
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: req.GetUsername(),
			}
			// set up the attribute of the task
			opts := []asynq.Option{
				asynq.MaxRetry(10),                // max retry per task
				asynq.ProcessIn(10 * time.Second), // processs after 10 seconds delay
				asynq.Queue(worker.QueueCritical), // send the task to a critical queue
			}
			err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
			return err
		},
	}

	// start a create user transaction
	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		// try to convert the error to a err.(*pq.Error) type
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exist %s", err)
			}
		}

		return nil, status.Errorf(codes.Internal, "failed create user %s", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}
	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := validation.ValidateUserName(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := validation.ValidateEmailAddress(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	if err := validation.ValidateFullName(req.GetFirstname(), req.GetLastname()); err != nil {
		violations = append(violations, fieldViolation("fullname", err))
	}
	return
}
