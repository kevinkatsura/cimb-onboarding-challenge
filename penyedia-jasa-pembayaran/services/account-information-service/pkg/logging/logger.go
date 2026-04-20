package logging

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func InitLogger() {
	l, _ := zap.NewProduction()
	logger = l.Sugar()
}

func Logger() *zap.SugaredLogger {
	return logger
}
