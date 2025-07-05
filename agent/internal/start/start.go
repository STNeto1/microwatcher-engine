package start

import (
	"context"
	"log/slog"
	"time"

	"github.com/microwatcher/agent/internal/config"
	"github.com/microwatcher/agent/internal/systeminformation"
)

func Start(ctx context.Context, config *config.Config) {
	aliveTicker := time.NewTicker(time.Second * 5)
	processTicker := time.NewTicker(config.Interval)

	defer aliveTicker.Stop()
	defer processTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-aliveTicker.C:
				config.Logger.Info("i'm alive")
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-processTicker.C:
				config.Logger.Info("processing info")
				runInfo := systeminformation.GetSystemInformation(config)
				config.Logger.Info(
					"done processing info",
					slog.Any("info", runInfo),
				)
			}
		}
	}()

	<-ctx.Done()
	config.Logger.Info("done running")
}
