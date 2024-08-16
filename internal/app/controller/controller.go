package controller

import "go.uber.org/zap"

var log = zap.Must(zap.NewProduction()).Sugar()
