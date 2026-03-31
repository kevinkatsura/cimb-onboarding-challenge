package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func InitLogger() (*zap.Logger, *zap.SugaredLogger, error) {
	// Ensure logs directory exists for promtail scraping
	os.MkdirAll("logs", 0755)

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", "logs/app.log"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

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
