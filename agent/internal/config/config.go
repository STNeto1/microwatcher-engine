package config

import (
	"log/slog"
	"os"
	"time"
)

const MinInterval = time.Second * 5

type Config struct {
	Logger     *slog.Logger
	Interval   time.Duration
	Identifier string
}

func NewConfig(logger *slog.Logger) *Config {
	return &Config{Logger: logger}
}

func (cfg *Config) SetInterval(val string) *Config {
	// TODO: maybe assert that val is not empty

	fromString, err := time.ParseDuration(val)
	if err != nil {
		cfg.Logger.Error("failed to parse duration",
			slog.String("value", val),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	cfg.Interval = fromString

	if fromString < MinInterval {
		cfg.Logger.Warn(
			"duration is too low",
			slog.String("value", val),
			slog.String("default value", MinInterval.String()),
		)
		cfg.Interval = MinInterval
	}

	return cfg
}

func (cfg *Config) SetIdentifier(val string) *Config {
	cfg.Identifier = val
	return cfg
}
