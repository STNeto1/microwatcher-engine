package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IngestClient struct {
	client v1.TelemetryServiceClient
	conn   *grpc.ClientConn
	Logger *slog.Logger
}

func NewIngestClient(addr string) *IngestClient {
	defaultLogger := logger.NewDefaultLogger()
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		defaultLogger.Error("failed to connect to ingest", slog.String("error", err.Error()))
		os.Exit(1)
	}

	client := v1.NewTelemetryServiceClient(conn)

	return &IngestClient{
		Logger: defaultLogger,
		client: client,
		conn:   conn,
	}
}

func (ic *IngestClient) SendData(telemetries []*v1.Telemetry) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*2,
	)
	defer cancel()

	response, err := ic.client.SendTelemetry(ctx, &v1.SendTelemetryRequest{
		Telemetries: telemetries,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("failed to send data"), err)
	}

	if !response.Success {
		return errors.New("failed to send data")
	}

	return nil
}

func (ic *IngestClient) HealthCheck(ctx context.Context, identifier string) error {
	ctx, cancel := context.WithTimeout(
		ctx,
		time.Second*2,
	)
	defer cancel()

	_, err := ic.client.HealthCheck(ctx, &v1.HealthCheckRequest{
		Timestamp:  timestamppb.Now(),
		Identifier: identifier,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("failed to health check"), err)
	}

	return nil
}

func (ic *IngestClient) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(
		ctx,
		time.Second*2,
	)
	defer cancel()

	_, err := ic.client.Ping(ctx, &v1.PingRequest{})
	if err != nil {
		return errors.Join(fmt.Errorf("failed to ping"), err)
	}

	return nil
}
