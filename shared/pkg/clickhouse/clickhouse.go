package clickhouse

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/microwatcher/shared/pkg/otlp"
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
	if err := chs.Conn.QueryRow(spanCtx, "SELECT id, label, secret FROM devices WHERE id = ? LIMIT 1", deviceID).ScanStruct(&device); err != nil {
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

	rows, err := chs.Conn.Query(spanCtx, "SELECT id, label, secret FROM devices")
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
