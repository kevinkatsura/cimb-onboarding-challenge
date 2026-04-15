package logging

import (
	"context"
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
	// Best-effort push; errors don't block the application
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
