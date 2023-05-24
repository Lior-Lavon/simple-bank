package gapi

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to Hash password %s", err)
	}

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		Firstname:      req.GetFirstname(),
		Lastname:       req.GetLastname(),
		Email:          req.GetEmail(),
	}

	user, err := server.store.CreateUser(ctx, arg)
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
		User: convertUser(user),
	}
	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	fmt.Println("-----------")
	fmt.Println("Username : ", req.Username)
	fmt.Println("FirstName : ", req.GetFirstname())
	fmt.Println("LastName : ", req.GetLastname())
	fmt.Println("Email : ", req.Email)
	fmt.Println("Password : ", req.Password)
	fmt.Println("-----------")

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
