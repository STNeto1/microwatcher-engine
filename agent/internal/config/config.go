package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/microwatcher/agent/internal/cli"
)

const MinInterval = time.Second * 5

type Config struct {
	Logger              *slog.Logger
	MetricInterval      time.Duration
	HealthCheckInterval time.Duration
	Identifier          string
	ClientID            string
	ClientSecret        []byte
}

func NewConfig(logger *slog.Logger) *Config {
	return &Config{Logger: logger}
}

func (cfg *Config) SetMetricInterval(val string) *Config {
	fromString, err := time.ParseDuration(val)
	if err != nil {
		cfg.Logger.Error("failed to parse duration",
			slog.String("value", val),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	if fromString < MinInterval {
		cfg.Logger.Warn(
			"duration is too low",
			slog.String("value", val),
			slog.String("default value", MinInterval.String()),
		)
		fromString = MinInterval
	}

	cfg.MetricInterval = fromString
	return cfg
}

func (cfg *Config) SetHealthCheckInterval(val string) *Config {
	fromString, err := time.ParseDuration(val)
	if err != nil {
		cfg.Logger.Error("failed to parse duration",
			slog.String("value", val),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	if fromString < MinInterval {
		cfg.Logger.Warn(
			"duration is too low",
			slog.String("value", val),
			slog.String("default value", MinInterval.String()),
		)
		fromString = MinInterval
	}

	cfg.HealthCheckInterval = fromString
	return cfg
}

func (cfg *Config) SetClientIDFromEnv() *Config {
	envClientID := os.Getenv("MW_CLIENT_ID")
	if envClientID != "" {
		cfg.ClientID = envClientID
	}
	return cfg
}

func (cfg *Config) SetClientID(val string) *Config {
	cfg.ClientID = val
	return cfg
}

func (cfg *Config) SetSecretFromEnv() *Config {
	envSecret := os.Getenv("MW_SECRET")
	if envSecret != "" {
		cfg.ClientSecret = []byte(envSecret)
	}
	return cfg
}

func (cfg *Config) SetSecret(val string) *Config {
	cfg.ClientSecret = []byte(val)
	return cfg
}

func (cfg *Config) SetDefaultIdentifier() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		cfg.Logger.Error("failed to get hostname", slog.Any("error", err))
		cfg.Identifier = "unknown"
	} else {
		cfg.Identifier = hostname
	}
	return cfg
}

func (cfg *Config) SetIdentifier(val string) *Config {
	cfg.Identifier = val
	return cfg
}

func (cfg *Config) ApplyStartOverrides(cliArgs cli.Start) *Config {
	cfg.
		SetMetricInterval(cliArgs.MetricInterval).
		SetHealthCheckInterval(cliArgs.HealthCheckInterval).
		SetDefaultIdentifier()

	if cliArgs.Identifier != "" {
		cfg.SetIdentifier(cliArgs.Identifier)
	}

	cfg.Logger.Info("applying overrides", slog.Any("cliArgs", cliArgs))

	cfg.SetClientIDFromEnv()
	if cliArgs.ClientID != "" {
		cfg.SetClientID(cliArgs.ClientID)
	}

	cfg.SetSecretFromEnv()
	if cliArgs.ClientSecret != "" {
		cfg.SetSecret(cliArgs.ClientSecret)
	}

	return cfg
}

func (cfg *Config) ApplyCheckOverrides(cliArgs cli.Check) *Config {
	cfg.
		SetDefaultIdentifier()

	cfg.SetClientIDFromEnv()
	if cliArgs.ClientID != "" {
		cfg.SetClientID(cliArgs.ClientID)
	}

	cfg.SetSecretFromEnv()
	if cliArgs.ClientSecret != "" {
		cfg.SetSecret(cliArgs.ClientSecret)
	}

	return cfg
}
