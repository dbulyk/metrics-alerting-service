package hashes

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func Hash(message, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil))
}
