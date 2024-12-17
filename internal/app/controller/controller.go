// Package controller
//
//	@Title			Shortener API
//	@Description	Сервис сокращения ссылок
//	@Version		1.0
package controller

import "go.uber.org/zap"

var log = zap.Must(zap.NewDevelopment()).Sugar()
