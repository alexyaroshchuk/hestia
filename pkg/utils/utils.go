package utils

import (
	"crypto/rand"
	"fmt"
	mranf "math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const defaultLength = 4

func RandStringBytes() string {
	b := make([]byte, defaultLength)
	for i := range b {
		b[i] = letterBytes[mranf.Intn(len(letterBytes))]
	}
	return string(b)
}

func TokenGenerator() (string, error) {
	b := make([]byte, defaultLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
