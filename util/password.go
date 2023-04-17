package util

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	// using random // different value every time
	//	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	//	return string(bytes), err

	// using SHA256 same value all the time
	s := password
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x\n", bs), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
