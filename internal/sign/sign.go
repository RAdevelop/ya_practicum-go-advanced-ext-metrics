package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 считает HMAC-SHA256 от строки и возвращает hex-строку
func SHA256(str []byte, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256Verify — сравнивает подписи для указанной строки и клюа
func SHA256Verify(str []byte, secretKey string, receivedHash string) bool {

	expectedHash := SHA256(str, secretKey)

	return hmac.Equal([]byte(receivedHash), []byte(expectedHash))
}
