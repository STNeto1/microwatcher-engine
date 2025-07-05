package internal

import (
	"context"
	"log/slog"

	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
)

type Server struct {
	Logger *slog.Logger
	v1.UnimplementedTelemetryServiceServer
}

func (s *Server) SendTelemetry(ctx context.Context, req *v1.SendTelemetryRequest) (*v1.SendTelemetryResponse, error) {
	s.Logger.Info(
		"SendTelemetry request received",
		slog.String("request", req.String()),
		slog.Any("telemetries", req.Telemetries),
	)

	return &v1.SendTelemetryResponse{Success: true}, nil
}
