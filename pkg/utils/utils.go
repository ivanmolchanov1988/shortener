package utils

import (
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
)

var allowedChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")

// RandStr генерирует рандомную строку.
func RandStr(n int) (string, error) {
	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		if err != nil {
			return "", err
		}
		b[i] = allowedChars[num.Int64()]
	}
	return string(b), nil
}

// GenUUID возвращает UID как строку.
func GenUUID() string {
	newUUID := uuid.New()
	return newUUID.String()
}
