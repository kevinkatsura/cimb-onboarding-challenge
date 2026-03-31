package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func InitLogger() (*zap.Logger, *zap.SugaredLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.Encoding = "console"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, nil, err
	}

	sugar := logger.Sugar()
	log = sugar
	return logger, sugar, nil
}

func Logger() *zap.SugaredLogger {
	if log == nil {
		_, sugar, _ := InitLogger()
		return sugar
	}
	return log
}
