package systeminformation

import (
	"time"

	"github.com/microwatcher/agent/internal/config"
	"github.com/shirou/gopsutil/mem"
)

type SystemInformation struct {
	Timestamp   time.Time
	TotalMemory uint64
	FreeMemory  uint64
	UsedMemory  uint64
}

func GetSystemInformation(cfg *config.Config) SystemInformation {
	v, _ := mem.VirtualMemory()
	// assert it doesn't fail

	return SystemInformation{
		Timestamp:   time.Now(),
		TotalMemory: v.Total,
		FreeMemory:  v.Free,
		UsedMemory:  v.Used,
	}
}
