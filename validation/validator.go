package validation

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUserName        = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString  // define regular expresion using 0-9, a-z, _ and each character can have multiple times
	isValidFullName        = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString // define regular expresion using a-z, [space - \\s] and each character can have multiple times
	isValidLowerCaseString = regexp.MustCompile(`^[a-z_]+$`).MatchString     // define regular expresion using a-z, _ and each character can have multiple times
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
	value := firstName + " " + lastName

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

func ValidateEmailId(value int64) error {
	// check lenght
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}

	return nil
}

func ValidateSecretCode(value string) error {
	// check lenght
	if err := ValidateString(value, 32, 128); err != nil {
		return err
	}

	// check using regular expressions
	if !isValidLowerCaseString(value) {
		return fmt.Errorf("must contain only lower case letters")
	}

	return nil
}
