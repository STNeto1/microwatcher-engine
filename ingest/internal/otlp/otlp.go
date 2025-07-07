package otlp

import (
	"context"
	"crypto/tls"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	serviceName    = "microwatcher-ingestor" // Name of the service for tracing.
	serviceVersion = "0.1.0"                 // Version of the service.
	otlpEndpoint   = "api.axiom.co"          // OTLP collector endpoint.
)

func createExporter(logger *slog.Logger) (sdktrace.SpanExporter, error) {
	if os.Getenv("AXIOM_TOKEN") == "" || os.Getenv("AXIOM_DATASET") == "" {
		logger.Info("No AXIOM_TOKEN or AXIOM_DATASET found, using stdout exporter",
			slog.String("env", "stdout"),
			slog.String("token", os.Getenv("AXIOM_TOKEN")),
			slog.String("dataset", os.Getenv("AXIOM_DATASET")),
		)
		return stdout.New(stdout.WithPrettyPrint())
	}

	logger.Info("Using AXIOM_TOKEN and AXIOM_DATASET")

	return otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":   os.Getenv("AXIOM_TOKEN"),
			"X-AXIOM-DATASET": os.Getenv("AXIOM_DATASET"),
		}),
		otlptracehttp.WithTLSClientConfig(&tls.Config{}),
	)
}

func InitLocalTracer(ctx context.Context, logger *slog.Logger) func() {
	exporter, err := createExporter(logger)
	if err != nil {
		logger.Error(
			"failed to create stdout exporter",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				semconv.ServiceVersionKey.String(serviceVersion),
				attribute.String("environment", os.Getenv("AXIOM_ENVIRONMENT")),
			),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		// Shutdown para garantir flush dos traces
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(
				"error shutting down tracer provider",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}
}
