package internal

import (
	"context"

	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
)

type Server struct {
	v1.UnimplementedTelemetryServiceServer
}

func (s *Server) SendTelemetry(ctx context.Context, req *v1.SendTelemetryRequest) (*v1.SendTelemetryResponse, error) {
	return &v1.SendTelemetryResponse{Success: true}, nil
}
