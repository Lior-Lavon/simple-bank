package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// //////////////////////////////////////
// PasetoMaker is a paseto token maker
type PasetoMaker struct {
	paseto *paseto.V2

	// user symertic-key algorithm
	symmetricKey []byte
}

// CreateToken create and sign a new token for a specific username and duration
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	// create new paylaod
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// return an encrypt token
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken checkes if the input token is valid or not
// if Valid , it will return the payload data stored inside the body of the token
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	// decode the token
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// check if the token is valid
	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// //////////////////////////////////////

// NewPasetoMaker create a new PasetoMaker instance
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	// check the symmetricKey len
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	// create a new PasetoMaker
	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}
