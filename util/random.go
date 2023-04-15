package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// Called automatically when the package is first generated
func init() {
	rand.Seed(time.Now().UnixNano())
}

// randomly generate an int between min & max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1) // return an int between min <-> max
}

// randonly returns a string of length n
func RandomString(n int) string {
	var ret strings.Builder

	for i := 0; i < n; i++ {
		pos := rand.Intn(len(alphabet))
		ret.WriteByte(alphabet[pos])
	}
	return ret.String()
}

// generate randome owner name
func RandomOwner() string {
	return RandomString(6) // random string of 6 letters
}

func RandEmail() string {
	name := RandomString(4)
	domain := RandomString(3)
	return name + "@" + domain + ".com"
}

// generate randome amount
func RandomMoney() int64 {
	return RandomInt(0, 1000) // generate a random int between x to y
}

// generate randome currency from a list
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	return currencies[rand.Intn(len(currencies))]
}

func GetTime() time.Time {
	return time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
}
