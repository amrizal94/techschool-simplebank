package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// randomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + r.Int63n(max-min+1)
}

// randomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[r.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrecy generates a random currency code
func RandomCurrecy() string {
	currencies := []string{IDR, USD, EUR, CAD}
	n := len(currencies)
	return currencies[r.Intn(n)]
}

func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(5))
}
