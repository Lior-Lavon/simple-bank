package validation

import (
	"fmt"
	"log"
	"net/mail"
	"regexp"
)

var (
	isValidUserName = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString // define regular expresion using 0-9, a-z, _ and each character can have multiple times
	isValidFullName = regexp.MustCompile(`^[a-zA-Z_]+$`).MatchString // define regular expresion using a-z, [space - \\s] and each character can have multiple times
)

// validate string length
func ValidateString(value string, minLength int, maxLength int) error {
	len := len(value)
	if len < minLength || len > maxLength {
		return fmt.Errorf("must contain from %d-%d characters", minLength, maxLength)
	}

	return nil
}

func ValidateUserName(value string) error {
	// check lenght
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	// check using regular expressions
	if !isValidUserName(value) {
		return fmt.Errorf("must contain only lowercase letters, numbers or underscor")
	}

	return nil
}

func ValidatePassword(value string) error {
	// check lenght
	return ValidateString(value, 6, 100)
}

func ValidateEmailAddress(value string) error {
	// check lenght
	if err := ValidateString(value, 3, 200); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is not a valida email address")
	}

	return nil
}

func ValidateFullName(firstName string, lastName string) error {
	value := firstName + "_" + lastName

	log.Println("value : ", value)

	// check lenght
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	// check using regular expressions
	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}

	return nil
}
