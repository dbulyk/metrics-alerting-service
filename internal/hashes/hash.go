package hashes

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	dst := mac.Sum(nil)
	sha := hex.EncodeToString(dst)
	return sha
}

func ValidHash(message, messageMAC, key string) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	sha, err := hex.DecodeString(messageMAC)
	if err != nil {
		return false
	}
	return hmac.Equal(sha, expectedMAC)
}
