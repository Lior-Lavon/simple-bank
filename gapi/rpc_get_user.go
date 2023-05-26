package gapi

import (
	"context"
	"database/sql"

	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {

	violations := validateGetUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed get user %s", err)
	}

	rsp := &pb.GetUserResponse{
		User: convertUser(user),
	}
	return rsp, nil
}

func validateGetUserRequest(req *pb.GetUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := validation.ValidateUserName(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	return
}
