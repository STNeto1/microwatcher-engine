package clickhouse

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ClickhouseSource struct {
	Conn driver.Conn
}

func (chs *ClickhouseSource) IngestV1MemoryTelemetries(ctx context.Context, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1MemoryTelemetries",
		trace.WithAttributes(),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO memory_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer batch.Close()

	for _, telemetry := range telemetries {
		_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
			trace.WithAttributes(
				attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
				attribute.String("identifier", telemetry.Identifier),
				attribute.Int64("total_memory", int64(telemetry.TotalMemory)),
				attribute.Int64("free_memory", int64(telemetry.FreeMemory)),
				attribute.Int64("used_memory", int64(telemetry.UsedMemory)),
			),
		)
		defer appendSpan.End()

		if err := batch.Append(
			telemetry.Timestamp.AsTime(),
			telemetry.Identifier,
			telemetry.TotalMemory,
			telemetry.FreeMemory,
			telemetry.UsedMemory,
		); err != nil {
			appendSpan.RecordError(err)
			appendSpan.SetStatus(codes.Error, "failed to append to batch")

			return errors.Join(errors.New("failed to append to batch"), err)
		}
	}

	if err := batch.Send(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to append to batch")

		return errors.Join(errors.New("failed to send batch"), err)
	}

	span.SetStatus(codes.Ok, "ingested")

	return nil
}

func (chs *ClickhouseSource) IngestV1CPUTelemetries(ctx context.Context, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1CPUTelemetries",
		trace.WithAttributes(),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO cpu_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer batch.Close()

	for _, telemetry := range telemetries {
		_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
			trace.WithAttributes(
				attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
				attribute.String("identifier", telemetry.Identifier),
				attribute.Float64("total_cpu", float64(telemetry.TotalCpu)),
				attribute.Float64("free_cpu", float64(telemetry.FreeCpu)),
				attribute.Float64("used_cpu", float64(telemetry.UsedCpu)),
			),
		)
		defer appendSpan.End()

		if err := batch.Append(
			telemetry.Timestamp.AsTime(),
			telemetry.Identifier,
			telemetry.TotalCpu,
			telemetry.FreeCpu,
			telemetry.UsedCpu,
		); err != nil {
			appendSpan.RecordError(err)
			appendSpan.SetStatus(codes.Error, "failed to append to batch")

			return errors.Join(errors.New("failed to append to batch"), err)
		}
	}

	if err := batch.Send(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to append to batch")

		return errors.Join(errors.New("failed to send batch"), err)
	}

	span.SetStatus(codes.Ok, "ingested")

	return nil
}

func NewLocalConnection(logger *slog.Logger) (*ClickhouseSource, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"192.168.1.7:9000"},
		Auth: clickhouse.Auth{
			Database: "microwatcher",
			Username: "default",
			Password: "",
		},
		Debug: true,
	})

	if err != nil {
		logger.Error("failed to connect to clickhouse",
			slog.String("error", err.Error()),
		)
		return nil, errors.Join(errors.New("failed to connect to clickhouse"), err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			logger.Error("failed to ping clickhouse",
				slog.Any("code", exception.Code),
				slog.String("message", exception.Message),
				slog.String("stacktrace", exception.StackTrace),
			)
		}
		return nil, errors.Join(errors.New("failed to ping clickhouse"), err)
	}

	return &ClickhouseSource{
		Conn: conn,
	}, nil
}
