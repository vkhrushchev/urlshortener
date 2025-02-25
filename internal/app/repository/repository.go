package repository

import (
	"errors"

	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// ErrConflict - короткая ссылка уже существует
// ErrNotFound - короткая ссылка не найдена
// ErrUnexpected - непредвиденная ошибка
var (
	ErrConflict   = errors.New("conflict")
	ErrNotFound   = errors.New("entity not found")
	ErrUnexpected = errors.New("unexpected error")
)
