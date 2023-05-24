package gapi

import (
	db "github.com/liorlavon/simplebank/db/sqlc"
	pb "github.com/liorlavon/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		Firstname:         user.Firstname,
		Lastname:          user.Lastname,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
