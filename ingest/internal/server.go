package internal

import (
	"context"
	"log/slog"

	"github.com/microwatcher/shared/pkg/clickhouse"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
)

type Server struct {
	Logger     *slog.Logger
	Clickhouse *clickhouse.ClickhouseSource
	v1.UnimplementedTelemetryServiceServer
}

func (svc *Server) SendTelemetry(ctx context.Context, req *v1.SendTelemetryRequest) (*v1.SendTelemetryResponse, error) {
	if err := svc.Clickhouse.IngestV1Telemetries(ctx, req.Telemetries); err != nil {
		svc.Logger.Error("failed to ingest telemetries",
			slog.String("error", err.Error()),
		)
		return &v1.SendTelemetryResponse{Success: false}, nil
	}

	svc.Logger.Info("telemetries ingested",
		slog.Int("size", len(req.Telemetries)),
	)

	return &v1.SendTelemetryResponse{Success: true}, nil
}
