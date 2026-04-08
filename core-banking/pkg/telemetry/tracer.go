package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Tracer trace.Tracer
var Meter metric.Meter

func init() {
	Tracer = noop.NewTracerProvider().Tracer("default-noop")
}

func InitProvider(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	serviceVersion := os.Getenv("SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", serviceName),
			attribute.String("service.version", serviceVersion),
			attribute.String("deployment.environment", environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// TRACER (OTLP gRPC to Tempo)
	tempoEndpoint := "tempo:4317"
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		tempoEndpoint = v
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(tempoEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	Tracer = tp.Tracer(serviceName)

	// METRICS (Prometheus Exporter)
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prometheus exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	Meter = mp.Meter(serviceName)

	return func(ctx context.Context) error {
		mpErr := mp.Shutdown(ctx)
		tpErr := tp.Shutdown(ctx)
		if mpErr != nil {
			return mpErr
		}
		return tpErr
	}, nil
}
