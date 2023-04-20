package token

import "time"

// Declear a Maker interface to create tokens, and later JWT & PASETO will implement this interface

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken create and sign a new token for a specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)

	// VerifyToken checkes if the input token is valid or not
	// if Valid , it will return the payload data stored inside the body of the token
	VerifyToken(token string) (*Payload, error)
}
