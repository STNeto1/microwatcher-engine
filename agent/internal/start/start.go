package start

import (
	"context"
	"log/slog"
	"time"

	"github.com/microwatcher/agent/internal"
	"github.com/microwatcher/agent/internal/config"
	"github.com/microwatcher/agent/internal/systeminformation"
	v1 "github.com/microwatcher/shared/pkg/gen/microwatcher/v1"
	"github.com/microwatcher/shared/pkg/iter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Start(ctx context.Context, config *config.Config) {
	aliveTicker := time.NewTicker(time.Second * 5)
	processTicker := time.NewTicker(config.MetricInterval)

	client := internal.NewIngestClient("localhost:50051")

	defer aliveTicker.Stop()
	defer processTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-aliveTicker.C:
				if err := client.HealthCheck(ctx, config.Identifier); err != nil {
					config.Logger.Error("failed to health check", slog.String("error", err.Error()))
					continue
				}
			}
		}
	}()

	go func() {
		var telemetries []*v1.Telemetry

		for {
			select {
			case <-ctx.Done():
				return
			case <-processTicker.C:
				runInfo := systeminformation.GetSystemInformation(config)

				telemetryDisks := iter.Map(runInfo.Disks, func(disk systeminformation.SystemInformationDisk) *v1.TelemetryDisk {
					return &v1.TelemetryDisk{
						Label:      disk.Label,
						Mountpoint: disk.Mountpoint,
						Total:      disk.Total,
						Used:       disk.Used,
						Free:       disk.Free,
					}
				})

				telemetryNetworks := iter.Map(runInfo.Networks, func(network systeminformation.SystemInformationNetwork) *v1.TelemetryNetwork {
					return &v1.TelemetryNetwork{
						Name:      network.Name,
						BytesSent: network.BytesSent,
						BytesRecv: network.BytesRecv,
					}
				})

				telemetries = append(telemetries, &v1.Telemetry{
					Timestamp:   timestamppb.Now(),
					Identifier:  config.Identifier,
					TotalMemory: runInfo.TotalMemory,
					FreeMemory:  runInfo.FreeMemory,
					UsedMemory:  runInfo.UsedMemory,
					TotalCpu:    runInfo.TotalCPU,
					FreeCpu:     runInfo.FreeCPU,
					UsedCpu:     runInfo.UsedCPU,
					Disks:       telemetryDisks,
					Networks:    telemetryNetworks,
				})

				if err := client.SendData(telemetries); err != nil {
					config.Logger.Error("failed to send data", slog.String("error", err.Error()))
					continue
				}

				telemetries = make([]*v1.Telemetry, 0)
			}
		}
	}()

	<-ctx.Done()
	config.Logger.Info("done running")
}
