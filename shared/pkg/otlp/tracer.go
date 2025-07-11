package otlp

import "go.opentelemetry.io/otel"

var IngestTracer = otel.Tracer("microwatcher-grpc")

var WebServerTracer = otel.Tracer("microwatcher-webserver")
