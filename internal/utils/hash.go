package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/rs/zerolog/log"
)

func Hash(message, key string) string {
	log.Info().Msgf("message: %s, key: %s", message, key)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return fmt.Sprintf("%x", h.Sum(nil))
}
