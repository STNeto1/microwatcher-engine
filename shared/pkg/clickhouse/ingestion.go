package clickhouse

import (
	"context"
	"errors"
	"log/slog"
	"time"

	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/otlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (chs *ClickhouseSource) IngestV1HealthCheck(ctx context.Context, deviceID string, healthcheck *v1.HealthCheckRequest) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1HealthCheck",
		trace.WithAttributes(
			attribute.String("timestamp", healthcheck.Timestamp.AsTime().Format(time.RFC3339)),
			attribute.String("identifier", healthcheck.Identifier),
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	if err := chs.Conn.Exec(spanCtx, "INSERT INTO health_checks (timestamp, device_id, identifier) values (?, ?, ?)",
		healthcheck.Timestamp.AsTime(),
		deviceID,
		healthcheck.Identifier,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}

	span.SetStatus(codes.Ok, "ingested")

	return nil
}

func (chs *ClickhouseSource) IngestV1MemoryTelemetries(ctx context.Context, deviceID string, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1MemoryTelemetries",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO memory_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer func() {
		if err := batch.Close(); err != nil {
			span.RecordError(err)
			chs.Logger.Error("failed to close batch",
				slog.String("error", err.Error()),
			)
		}
	}()

	for _, telemetry := range telemetries {
		_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
			trace.WithAttributes(
				attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
				attribute.String("deviceID", deviceID),
				attribute.String("identifier", telemetry.Identifier),
				attribute.Int64("total_memory", int64(telemetry.TotalMemory)),
				attribute.Int64("free_memory", int64(telemetry.FreeMemory)),
				attribute.Int64("used_memory", int64(telemetry.UsedMemory)),
			),
		)
		defer appendSpan.End()

		if err := batch.Append(
			telemetry.Timestamp.AsTime(),
			deviceID,
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

func (chs *ClickhouseSource) IngestV1CPUTelemetries(ctx context.Context, deviceID string, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1CPUTelemetries",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO cpu_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer func() {
		if err := batch.Close(); err != nil {
			span.RecordError(err)
			chs.Logger.Error("failed to close batch",
				slog.String("error", err.Error()),
			)
		}
	}()

	for _, telemetry := range telemetries {
		_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
			trace.WithAttributes(
				attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
				attribute.String("deviceID", deviceID),
				attribute.String("identifier", telemetry.Identifier),
				attribute.Float64("total_cpu", float64(telemetry.TotalCpu)),
				attribute.Float64("free_cpu", float64(telemetry.FreeCpu)),
				attribute.Float64("used_cpu", float64(telemetry.UsedCpu)),
			),
		)
		defer appendSpan.End()

		if err := batch.Append(
			telemetry.Timestamp.AsTime(),
			deviceID,
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

func (chs *ClickhouseSource) IngestV1DisksTelemetries(ctx context.Context, deviceID string, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1DisksTelemetries",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO disk_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer func() {
		if err := batch.Close(); err != nil {
			span.RecordError(err)
			chs.Logger.Error("failed to close batch",
				slog.String("error", err.Error()),
			)
		}
	}()

	for _, telemetry := range telemetries {
		for _, disk := range telemetry.Disks {
			_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
				trace.WithAttributes(
					attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
					attribute.String("deviceID", deviceID),
					attribute.String("identifier", telemetry.Identifier),
					attribute.String("label", disk.Label),
					attribute.String("mountpoint", disk.Mountpoint),
					attribute.Int64("total_disk", int64(disk.Total)),
					attribute.Int64("free_disk", int64(disk.Free)),
					attribute.Int64("used_disk", int64(disk.Used)),
				),
			)
			defer appendSpan.End()

			if err := batch.Append(
				telemetry.Timestamp.AsTime(),
				deviceID,
				telemetry.Identifier,
				disk.Label,
				disk.Mountpoint,
				disk.Total,
				disk.Free,
				disk.Used,
			); err != nil {
				appendSpan.RecordError(err)
				appendSpan.SetStatus(codes.Error, "failed to append to batch")

				return errors.Join(errors.New("failed to append to batch"), err)
			}
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

func (chs *ClickhouseSource) IngestV1NetworksTelemetries(ctx context.Context, deviceID string, telemetries []*v1.Telemetry) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.IngestV1NetworksTelemetries",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	batch, err := chs.Conn.PrepareBatch(spanCtx, "INSERT INTO network_telemetries")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to prepare batch")

		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer func() {
		if err := batch.Close(); err != nil {
			span.RecordError(err)
			chs.Logger.Error("failed to close batch",
				slog.String("error", err.Error()),
			)
		}
	}()

	for _, telemetry := range telemetries {
		for _, nic := range telemetry.Networks {
			_, appendSpan := otlp.IngestTracer.Start(ctx, "appending to batch",
				trace.WithAttributes(
					attribute.String("timestamp", telemetry.Timestamp.AsTime().Format(time.RFC3339)),
					attribute.String("deviceID", deviceID),
					attribute.String("identifier", telemetry.Identifier),
					attribute.String("name", nic.Name),
					attribute.Int64("bytes_sent", int64(nic.BytesSent)),
					attribute.Int64("bytes_received", int64(nic.BytesRecv)),
				),
			)
			defer appendSpan.End()

			if err := batch.Append(
				telemetry.Timestamp.AsTime(),
				deviceID,
				telemetry.Identifier,
				nic.Name,
				nic.BytesSent,
				nic.BytesRecv,
			); err != nil {
				appendSpan.RecordError(err)
				appendSpan.SetStatus(codes.Error, "failed to append to batch")

				return errors.Join(errors.New("failed to append to batch"), err)
			}
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
