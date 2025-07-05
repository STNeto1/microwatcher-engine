package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/microwatcher/ingest/internal"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/logger"

	"google.golang.org/grpc"
)

const Port = "50051"

func main() {
	logger := logger.NewDefaultLogger()

	lis, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		logger.Error(
			"failed to listen",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	s := grpc.NewServer()
	v1.RegisterTelemetryServiceServer(s, &internal.Server{})

	logger.Info("Starting server...", slog.String("port", Port))

	if err := s.Serve(lis); err != nil {
		logger.Error(
			"failed to serve",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}
