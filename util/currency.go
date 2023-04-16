package util

// implement the login to check if the currency is supported or not
// list of supported currency
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// return true if currency supported
func IsSuportedCurrency(currency string) bool {

	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
