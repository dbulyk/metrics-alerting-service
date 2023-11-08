package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/rs/zerolog/log"
)

func Hash(message, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	log.Info().Msgf("message: %s, key: %s, hash: %x", message, key, h.Sum(nil))
	return fmt.Sprintf("%x", h.Sum(nil))
}
