package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func InitLogger() (*zap.Logger, *zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	lokiURL := "http://loki:3100/loki/api/v1/push"
	// Fallback for local testing outside of docker
	if os.Getenv("LOKI_URL") != "" {
		lokiURL = os.Getenv("LOKI_URL")
	}

	lokiSyncer := NewLokiSyncer(lokiURL, map[string]string{
		"job": "core-banking-logs",
	})
	
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), lokiSyncer),
		config.Level,
	)

	logger := zap.New(core, zap.AddCaller())
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
