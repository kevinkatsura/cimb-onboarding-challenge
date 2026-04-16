package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// LokiSyncer pushes log lines to Grafana Loki via HTTP push API.
type LokiSyncer struct {
	url    string
	labels map[string]string
	client *http.Client
}

func NewLokiSyncer(url string, labels map[string]string) *LokiSyncer {
	return &LokiSyncer{url: url, labels: labels, client: &http.Client{Timeout: 5 * time.Second}}
}

func (l *LokiSyncer) Write(p []byte) (int, error) {
	// Prepare Loki push payload
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("loki push failed with status: %d", resp.StatusCode)
	}

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

	lokiSyncer := NewLokiSyncer(lokiURL, map[string]string{"job": "core-banking-system"})

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

// Ctx returns a logger enriched with OpenTelemetry trace_id and span_id for Loki correlation.
func Ctx(ctx context.Context) *zap.SugaredLogger {
	l := Logger()
	if ctx == nil {
		return l
	}
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return l.With(
			"trace_id", spanCtx.TraceID().String(),
			"span_id", spanCtx.SpanID().String(),
		)
	}
	return l
}
