package internal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/microwatcher/agent/internal/config"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IngestClient struct {
	client       v1.TelemetryServiceClient
	conn         *grpc.ClientConn
	Logger       *slog.Logger
	ClientID     string
	ClientSecret []byte
}

func NewIngestClient(addr string, cfg *config.Config) *IngestClient {
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
		Logger:       defaultLogger,
		client:       client,
		conn:         conn,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}
}

func (ic *IngestClient) SendData(telemetries []*v1.Telemetry) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*2,
	)
	defer cancel()

	req := &v1.SendTelemetryRequest{
		Telemetries: telemetries,
	}

	payloadBytes, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	mac := hmac.New(sha256.New, ic.ClientSecret)
	mac.Write(payloadBytes)
	signature := mac.Sum(nil)
	signatureHex := hex.EncodeToString(signature)

	md := metadata.New(map[string]string{
		"x-signature": signatureHex,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := ic.client.SendTelemetry(ctx, req)
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

	req := &v1.PingRequest{}

	payloadBytes, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	mac := hmac.New(sha256.New, ic.ClientSecret)
	mac.Write(payloadBytes)
	signature := mac.Sum(nil)
	signatureHex := hex.EncodeToString(signature)

	md := metadata.New(map[string]string{
		"x-signature": signatureHex,
		"x-client-id": ic.ClientID,
	})

	_, err = ic.client.Ping(metadata.NewOutgoingContext(ctx, md), req)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to ping"), err)
	}

	return nil
}
