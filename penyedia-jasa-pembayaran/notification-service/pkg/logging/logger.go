package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

type LokiSyncer struct {
	url    string
	labels map[string]string
	client *http.Client
}

func NewLokiSyncer(url string, labels map[string]string) *LokiSyncer {
	return &LokiSyncer{url: url, labels: labels, client: &http.Client{Timeout: 5 * time.Second}}
}

func (l *LokiSyncer) Write(p []byte) (int, error) {
	now := time.Now().UnixNano()
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": l.labels,
				"values": [][]string{
					{fmt.Sprintf("%d", now), string(p)},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	resp, err := l.client.Post(l.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return len(p), nil
}

func (l *LokiSyncer) Sync() error { return nil }

func InitLogger() (*zap.Logger, *zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	lokiURL := "http://loki:3100/loki/api/v1/push"
	if os.Getenv("LOKI_URL") != "" {
		lokiURL = os.Getenv("LOKI_URL")
	}

	lokiSyncer := NewLokiSyncer(lokiURL, map[string]string{"job": "notification-service"})

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), lokiSyncer),
		cfg.Level,
	)

	logger := zap.New(core, zap.AddCaller())
	sugar := logger.Sugar()
	log = sugar
	return logger, sugar, nil
}

func Logger() *zap.SugaredLogger {
	if log == nil {
		_, s, _ := InitLogger()
		return s
	}
	return log
}
