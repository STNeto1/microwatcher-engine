package internal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/microwatcher/shared/pkg/clickhouse"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/otlp"
	"github.com/samborkent/uuidv7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	Logger     *slog.Logger
	Clickhouse *clickhouse.ClickhouseSource
	v1.UnimplementedTelemetryServiceServer
}

func extractHeader(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (svc *Server) ValidateMetadata(ctx context.Context, md metadata.MD) (string, string, error) {
	signature := extractHeader(md, "x-signature")
	clientID := extractHeader(md, "x-client-id")

	if signature == "" || clientID == "" {
		svc.Logger.Error("missing credentials from request",
			slog.String("signature", signature),
			slog.String("clientID", clientID),
		)

		return "", "", fmt.Errorf("missing credentials")
	}

	if valid := uuidv7.IsValidString(clientID); !valid {
		svc.Logger.Error("invalid client id",
			slog.String("clientID", clientID),
		)

		// span.RecordError(fmt.Errorf("invalid client id"))
		// span.SetStatus(codes.Error, "invalid client id")
		return "", "", fmt.Errorf("invalid client id")
	}

	return signature, clientID, nil
}

func (svc *Server) ValidateSignature(ctx context.Context, signature string, deviceID string, msg proto.Message) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "Server.ValidateSignature",
		trace.WithAttributes(
			attribute.String("signature", signature),
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	payloadBytes, err := proto.Marshal(msg)
	if err != nil {
		svc.Logger.Error("failed to marshal request",
			slog.String("error", err.Error()),
		)

		return errors.Join(errors.New("failed to marshal"), err)
	}

	deviceInfo, err := svc.Clickhouse.FindDeviceByID(spanCtx, deviceID)
	if err != nil {
		svc.Logger.Error("failed to find client",
			slog.String("error", err.Error()),
		)

		span.RecordError(fmt.Errorf("failed to find client"))
		span.SetStatus(codes.Error, "failed to find client")
		return errors.Join(errors.New("failed to find client"), err)
	}

	mac := hmac.New(sha256.New, []byte(deviceInfo.Secret))
	mac.Write(payloadBytes)
	expectedSig := mac.Sum(nil)
	expectedSigHex := hex.EncodeToString(expectedSig)

	if !hmac.Equal([]byte(signature), []byte(expectedSigHex)) {
		svc.Logger.Error("invalid signature",
			slog.String("signature", signature),
			slog.String("expected", expectedSigHex),
		)

		span.RecordError(fmt.Errorf("invalid signature"))
		span.SetStatus(codes.Error, "invalid signature")
		return fmt.Errorf("invalid signature")
	}

	span.SetStatus(codes.Ok, "validated payload")

	return nil
}

func (svc *Server) Ping(ctx context.Context, req *v1.PingRequest) (*v1.PingResponse, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "Server.Ping",
		trace.WithAttributes(attribute.String("method", "Ping")),
		trace.WithAttributes(),
	)
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		svc.Logger.Error("failed to get metadata")

		span.RecordError(fmt.Errorf("failed to get metadata"))
		span.SetStatus(codes.Error, "failed to get metadata")
		return nil, fmt.Errorf("failed to get metadata")
	}

	signature, deviceID, err := svc.ValidateMetadata(ctx, md)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := svc.ValidateSignature(spanCtx, signature, deviceID, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.Join(errors.New("failed to validate signature"), err)
	}

	svc.Logger.Info("ping received",
		slog.Any("signature", signature),
		slog.Any("deviceID", deviceID),
	)

	return &v1.PingResponse{}, nil
}

func (svc *Server) HealthCheck(ctx context.Context, req *v1.HealthCheckRequest) (*v1.Empty, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "Server.HealthCheck",
		trace.WithAttributes(attribute.String("method", "HealthCheck")),
		trace.WithAttributes(),
	)
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		svc.Logger.Error("failed to get metadata")

		span.RecordError(fmt.Errorf("failed to get metadata"))
		span.SetStatus(codes.Error, "failed to get metadata")
		return nil, fmt.Errorf("failed to get metadata")
	}

	signature, deviceID, err := svc.ValidateMetadata(ctx, md)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := svc.ValidateSignature(spanCtx, signature, deviceID, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.Join(errors.New("failed to validate signature"), err)
	}

	// TODO: maybe retry or send to a "dead" queue to retry later
	if err := svc.Clickhouse.IngestV1HealthCheck(spanCtx, req); err != nil {
		svc.Logger.Error("failed to ingest health check",
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ingest health check")
		return &v1.Empty{}, nil
	}

	return &v1.Empty{}, nil
}

func (svc *Server) SendTelemetry(ctx context.Context, req *v1.SendTelemetryRequest) (*v1.SendTelemetryResponse, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "Server.SendTelemetry",
		trace.WithAttributes(attribute.String("method", "SendTelemetry")),
		trace.WithAttributes(attribute.Int("batch size", len(req.Telemetries))),
	)
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		svc.Logger.Error("failed to get metadata")

		span.RecordError(fmt.Errorf("failed to get metadata"))
		span.SetStatus(codes.Error, "failed to get metadata")
		return nil, fmt.Errorf("failed to get metadata")
	}

	signature, deviceID, err := svc.ValidateMetadata(ctx, md)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := svc.ValidateSignature(spanCtx, signature, deviceID, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.Join(errors.New("failed to validate signature"), err)
	}

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

	if err := svc.Clickhouse.IngestV1DisksTelemetries(spanCtx, req.Telemetries); err != nil {
		svc.Logger.Error("failed to ingest disks telemetries",
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ingest telemetries")
		return &v1.SendTelemetryResponse{Success: false}, nil
	}

	if err := svc.Clickhouse.IngestV1NetworksTelemetries(spanCtx, req.Telemetries); err != nil {
		svc.Logger.Error("failed to ingest networks telemetries",
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
