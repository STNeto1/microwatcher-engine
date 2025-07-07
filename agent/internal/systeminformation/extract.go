package systeminformation

import (
	"time"

	"github.com/microwatcher/agent/internal/config"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type SystemInformationDisk struct {
	Label      string
	Mountpoint string
	Total      uint64
	Free       uint64
	Used       uint64
}

type SystemInformation struct {
	Timestamp   time.Time
	TotalMemory uint64
	FreeMemory  uint64
	UsedMemory  uint64
	TotalCPU    float32
	FreeCPU     float32
	UsedCPU     float32
	Disks       []SystemInformationDisk
}

func GetSystemInformation(cfg *config.Config) SystemInformation {
	v, _ := mem.VirtualMemory()
	// TODO: assert it doesn't fail

	stats, _ := cpu.Times(false)
	// TODO: assert it doesn't fail

	currStats := stats[0]
	totalCPU := currStats.User + currStats.System + currStats.Idle + currStats.Nice +
		currStats.Iowait + currStats.Irq + currStats.Softirq + currStats.Steal +
		currStats.Guest + currStats.GuestNice
	freeCPU := currStats.Idle + currStats.Iowait
	usedCPU := totalCPU - freeCPU

	parts, _ := disk.Partitions(true)
	// TODO: assert it doesn't fail

	systemDisks := make([]SystemInformationDisk, len(parts))
	for i, part := range parts {
		usage, _ := disk.Usage(part.Mountpoint)
		// TODO: assert it doesn't fail

		systemDisks[i] = SystemInformationDisk{
			Label:      part.Device,
			Mountpoint: part.Mountpoint,
			Total:      usage.Total,
			Free:       usage.Free,
			Used:       usage.Used,
		}
	}

	return SystemInformation{
		Timestamp:   time.Now(),
		TotalMemory: v.Total,
		FreeMemory:  v.Free,
		UsedMemory:  v.Used,
		TotalCPU:    float32(totalCPU),
		FreeCPU:     float32(freeCPU),
		UsedCPU:     float32(usedCPU),
		Disks:       systemDisks,
	}
}
