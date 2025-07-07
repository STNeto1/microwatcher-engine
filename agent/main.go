package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/microwatcher/agent/internal/cli"
	"github.com/microwatcher/agent/internal/config"
	"github.com/microwatcher/agent/internal/start"

	"github.com/alecthomas/kong"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var cliArgs cli.CLI

	kongCtx := kong.Parse(&cliArgs,
		kong.Name("mw-agent"),
		kong.Description("Agent for microwatcher"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	agentConfig := config.
		NewConfig(jsonLogger).
		SetMetricInterval(cliArgs.Start.MetricInterval).
		SetHealthCheckInterval(cliArgs.Start.HealthCheckInterval).
		SetIdentifier(cliArgs.Start.Identifier)

	switch kongCtx.Command() {
	case "check":
		start.Ping(ctx, agentConfig)
	case "start":
		start.Start(ctx, agentConfig)
	default:
		panic(kongCtx.Command())
	}

}
