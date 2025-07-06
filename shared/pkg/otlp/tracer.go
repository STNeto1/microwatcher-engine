package otlp

import "go.opentelemetry.io/otel"

var IngestTracer = otel.Tracer("microwatcher-grpc")

// var Tracer = otel.Tracer("microwatcher-grpc")
