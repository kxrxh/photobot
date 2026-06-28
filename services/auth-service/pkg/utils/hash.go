package utils

import (
	"crypto/hmac"
	"crypto/sha256"
)

func HmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
