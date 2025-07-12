package clickhouse

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/microwatcher/shared/pkg/otlp"
	"github.com/microwatcher/shared/pkg/utils"
	"github.com/samborkent/uuidv7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ClickhouseSource struct {
	Conn   driver.Conn
	Logger *slog.Logger
}

func (chs *ClickhouseSource) FindDeviceByID(ctx context.Context, deviceID string) (*ClickhouseDevice, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.FindDeviceByID",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	var device ClickhouseDevice
	if err := chs.Conn.QueryRow(spanCtx, "SELECT id, label, secret, version FROM devices FINAL WHERE id = ? LIMIT 1", deviceID).ScanStruct(&device); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		return nil, errors.Join(errors.New("failed to query"), err)
	}

	span.SetStatus(codes.Ok, "found device")

	return &device, nil
}

func (chs *ClickhouseSource) ListDevices(ctx context.Context) ([]*ClickhouseDevice, error) {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.ListDevices",
		trace.WithAttributes(),
	)

	defer span.End()

	rows, err := chs.Conn.Query(spanCtx, "SELECT id, label, secret FROM devices FINAL")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		return nil, errors.Join(errors.New("failed to query"), err)
	}

	var devices []*ClickhouseDevice
	for rows.Next() {
		var d ClickhouseDevice
		if err := rows.ScanStruct(&d); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to scan")

			return nil, errors.Join(errors.New("failed to scan"), err)
		}

		devices = append(devices, &d)
	}

	span.SetStatus(codes.Ok, "listed devices")

	return devices, nil
}

func (chs *ClickhouseSource) CreateDevice(ctx context.Context, label string) (*ClickhouseDevice, error) {

	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.CreateDevice",
		trace.WithAttributes(
			attribute.String("label", label),
		),
	)
	defer span.End()

	genUUID := uuidv7.New()

	record := ClickhouseDevice{
		ID:     uuid.MustParse(genUUID.String()),
		Label:  label,
		Secret: utils.RandomString(32),
	}

	if err := chs.Conn.Exec(spanCtx, "INSERT INTO devices (id, label, secret) values (?, ?, ?)",
		record.ID,
		record.Label,
		record.Secret,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create device")

		return nil, errors.Join(errors.New("failed to create device"), err)
	}

	return &record, nil
}

func (chs *ClickhouseSource) ResetDeviceSecret(ctx context.Context, deviceID string) error {
	spanCtx, span := otlp.IngestTracer.Start(ctx, "ClickhouseSource.ResetDeviceSecret",
		trace.WithAttributes(
			attribute.String("deviceID", deviceID),
		),
	)
	defer span.End()

	existingDevice, err := chs.FindDeviceByID(spanCtx, deviceID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to find device")

		return errors.Join(errors.New("failed to find device"), err)
	}

	if err := chs.Conn.Exec(spanCtx, "INSERT into devices (id, label, secret, version) VALUES (?, ?, ?, ?)",
		deviceID,
		existingDevice.Label,
		utils.RandomString(32),
		existingDevice.Version+1,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update device")

		return errors.Join(errors.New("failed to update device"), err)
	}

	return nil
}
