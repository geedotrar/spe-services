package helpers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
)

func SignHMACSHA512(payload, key string) string {
	hash := hmac.New(sha512.New, []byte(key))
	hash.Write([]byte(payload))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
