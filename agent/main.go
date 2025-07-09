package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/alecthomas/kong"
	"github.com/microwatcher/agent/internal/cli"
	"github.com/microwatcher/agent/internal/config"
	"github.com/microwatcher/agent/internal/start"
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
		NewConfig(jsonLogger)

	switch kongCtx.Command() {
	case "check":
		agentConfig.ApplyCheckOverrides(cliArgs.Check)
		start.Ping(ctx, agentConfig)
	case "start":
		agentConfig.ApplyStartOverrides(cliArgs.Start)
		start.Start(ctx, agentConfig)
	default:
		panic(kongCtx.Command())
	}
}
