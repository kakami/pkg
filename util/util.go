package util

import (
	"math/rand"
	"time"
)

var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_")

func RandomString(length int) string {
	var out []byte
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		out = append(out, chars[rand.Int()%len(chars)])
	}
	return string(out)
}
