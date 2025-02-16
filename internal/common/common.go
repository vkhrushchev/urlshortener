package common

import (
	"crypto/md5"
	"encoding/hex"
)

// StringContextKey тип для ключей контекста приложения Shortener
type StringContextKey string

// UserIDContextKey - ключ для хранения идентификатора пользователя
const (
	UserIDContextKey StringContextKey = "userID"
)

func CheckSignature(toCheck, toCheckSignature, salt string) bool {
	toCheckExpectedSignatureBytes := md5.Sum([]byte(toCheck + salt))
	toCheckExpectedSignature := hex.EncodeToString(toCheckExpectedSignatureBytes[:])

	return toCheckExpectedSignature == toCheckSignature
}
