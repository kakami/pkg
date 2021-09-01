package util

import (
	"math"
	"math/rand"
)

var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")

func RandomString(length int) string {
	var out []byte
	// rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		out = append(out, chars[rand.Int()%len(chars)])
	}
	return string(out)
}

func Prime(d int64) int64 {
	for i := d; ; i++ {
		if isPrime(i) {
			return i
		}
	}
}

func isPrime(d int64) bool {
	if d == 2 || d == 3 {
		return true
	}
	if d < 2 || d%2 == 0 || d%3 == 0 {
		return false
	}
	sd := int64(math.Sqrt(float64(d)))
	for i := int64(5); i <= sd; i += 2 {
		if d%i == 0 {
			return false
		}
	}
	return true
}
