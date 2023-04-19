package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/liorlavon/simplebank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {

	// get the value of the field
	currency, ok := fieldLevel.Field().Interface().(string)
	if ok {
		// then currency is a valid string

		// check if this currency is supported
		return util.IsSuportedCurrency(currency)
	}
	// the currency field is not a string
	return false
}

/*
var validEmail validator.Func = func(fl validator.FieldLevel) bool {
	// get the filed email and check if it is a string after conversion
	email, ok := fl.Field().Interface().(string)
	if ok {

		// check if the email is valid
		return util.IsEmailValid(email)
	}
	return false
}
*/
