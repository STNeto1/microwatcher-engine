package systeminformation

import (
	"time"

	"github.com/microwatcher/agent/internal/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type SystemInformation struct {
	Timestamp   time.Time
	TotalMemory uint64
	FreeMemory  uint64
	UsedMemory  uint64
	TotalCPU    float32
	FreeCPU     float32
	UsedCPU     float32
}

func GetSystemInformation(cfg *config.Config) SystemInformation {
	v, _ := mem.VirtualMemory()
	// assert it doesn't fail

	stats, _ := cpu.Times(false)
	// assert it doesn't fail

	currStats := stats[0]
	totalCPU := currStats.User + currStats.System + currStats.Idle + currStats.Nice +
		currStats.Iowait + currStats.Irq + currStats.Softirq + currStats.Steal +
		currStats.Guest + currStats.GuestNice
	freeCPU := currStats.Idle + currStats.Iowait
	usedCPU := totalCPU - freeCPU

	return SystemInformation{
		Timestamp:   time.Now(),
		TotalMemory: v.Total,
		FreeMemory:  v.Free,
		UsedMemory:  v.Used,
		TotalCPU:    float32(totalCPU),
		FreeCPU:     float32(freeCPU),
		UsedCPU:     float32(usedCPU),
	}
}
