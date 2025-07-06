package clickhouse

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
)

type ClickhouseSource struct {
	Conn driver.Conn
}

func (chs *ClickhouseSource) IngestV1Telemetries(ctx context.Context, telemetries []*v1.Telemetry) error {
	batch, err := chs.Conn.PrepareBatch(ctx, "INSERT INTO memory_telemetries")
	if err != nil {
		return errors.Join(errors.New("failed to prepare batch"), err)
	}
	defer batch.Close()

	for _, telemetry := range telemetries {
		if err := batch.Append(
			telemetry.Timestamp.AsTime(),
			telemetry.Identifier,
			telemetry.TotalMemory,
			telemetry.FreeMemory,
			telemetry.UsedMemory,
		); err != nil {
			return errors.Join(errors.New("failed to append to batch"), err)
		}
	}

	if err := batch.Send(); err != nil {
		return errors.Join(errors.New("failed to send batch"), err)
	}

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
