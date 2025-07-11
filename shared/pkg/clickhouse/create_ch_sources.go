package clickhouse

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func NewLocalConnection(logger *slog.Logger) (*ClickhouseSource, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"192.168.1.7:9000"},
		Auth: clickhouse.Auth{
			Database: "microwatcher",
			Username: "default",
			Password: "",
		},
		Debug: false,
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
		Conn:   conn,
		Logger: logger,
	}, nil
}
