package internal

import (
	"context"
	"log/slog"

	"github.com/microwatcher/shared/pkg/clickhouse"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	Logger     *slog.Logger
	Clickhouse *clickhouse.ClickhouseSource
	v1.UnimplementedTelemetryServiceServer
}

func (svc *Server) SendTelemetry(ctx context.Context, req *v1.SendTelemetryRequest) (*v1.SendTelemetryResponse, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "Server.SendTelemetry",
		trace.WithAttributes(attribute.String("method", "SendTelemetry")),
		trace.WithAttributes(attribute.Int("batch size", len(req.Telemetries))),
	)
	defer span.End()

	// TODO: maybe retry or send to a "dead" queue to retry later
	if err := svc.Clickhouse.IngestV1MemoryTelemetries(spanCtx, req.Telemetries); err != nil {
		svc.Logger.Error("failed to ingest memory telemetries",
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ingest telemetries")
		return &v1.SendTelemetryResponse{Success: false}, nil
	}

	if err := svc.Clickhouse.IngestV1CPUTelemetries(spanCtx, req.Telemetries); err != nil {
		svc.Logger.Error("failed to ingest cpu telemetries",
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ingest telemetries")
		return &v1.SendTelemetryResponse{Success: false}, nil
	}

	svc.Logger.Info("telemetries ingested",
		slog.Int("size", len(req.Telemetries)),
	)
	span.SetStatus(codes.Ok, "ingested")

	return &v1.SendTelemetryResponse{Success: true}, nil
}
