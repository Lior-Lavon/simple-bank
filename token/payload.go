package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Different type of error returned by the VerifyToken function
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

// ///////////////////
// Payloadcontain the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`         // handle to the token in case we need to expire the token
	Username  string    `json:"username"`   // username of the user
	IssuedAt  time.Time `json:"issued_at"`  // when the token was created
	ExpiredAt time.Time `json:"expired_at"` // when the token will expire
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}

/////////////////////

// NewPayload will create new token payload with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}
