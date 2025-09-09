package metrics

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ResourcePolling struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
}

func getContainerCPUPercent(interval time.Duration) (float64, error) {
	quota, err := readInt64("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if err != nil {
		return 0, err
	}

	period, err := readInt64("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err != nil {
		return 0, err
	}

	usage1, err := readInt64("/sys/fs/cgroup/cpuacct/cpuacct.usage")
	if err != nil {
		return 0, err
	}

	time.Sleep(interval)

	usage2, err := readInt64("/sys/fs/cgroup/cpuacct/cpuacct.usage")
	if err != nil {
		return 0, err
	}

	maxCores := float64(quota) / float64(period)
	deltaUsageSec := float64(usage2-usage1) / 1_000_000_000
	percent := (deltaUsageSec / interval.Seconds()) / maxCores * 100

	return percent, nil
}

func getMemoryUsagePercent() (float64, error) {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0, err
	}

	total := float64(info.Totalram * uint64(info.Unit))
	free := float64(info.Freeram * uint64(info.Unit))

	return (total - free) / total * 100, nil
}

func getDiskUsagePercent(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	total := float64(stat.Blocks * uint64(stat.Bsize))
	available := float64(stat.Bavail * uint64(stat.Bsize))

	return (total - available) / total * 100, nil
}

func getSystemCPUPercent() (float64, error) {
	getCPUTimes := func() (idle, total int64, err error) {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return 0, 0, err
		}

		line := strings.Split(string(data), "\n")[0]
		fields := strings.Fields(line)[1:] // skip "cpu" label

		for i, field := range fields {
			val, err := strconv.ParseInt(field, 10, 64)
			if err != nil {
				return 0, 0, err
			}
			total += val
			if i == 3 { // idle is 4th field (index 3)
				idle = val
			}
		}
		return idle, total, nil
	}

	idle1, total1, err := getCPUTimes()
	if err != nil {
		return 0, err
	}

	time.Sleep(100 * time.Millisecond)

	idle2, total2, err := getCPUTimes()
	if err != nil {
		return 0, err
	}

	idleDelta := idle2 - idle1
	totalDelta := total2 - total1

	if totalDelta == 0 {
		return 0, nil
	}

	return (1.0 - float64(idleDelta)/float64(totalDelta)) * 100, nil
}

func GetContainerAvailableResources() (ResourcePolling, error) {
	cpu, err := getContainerCPUPercent(time.Second)
	if err != nil {
		return ResourcePolling{}, err
	}

	memory, err := getMemoryUsagePercent()
	if err != nil {
		return ResourcePolling{}, err
	}

	disk, err := getDiskUsagePercent("/")
	if err != nil {
		return ResourcePolling{}, err
	}

	return ResourcePolling{
		CPUUsagePercent:    cpu,
		MemoryUsagePercent: memory,
		DiskUsagePercent:   disk,
	}, nil
}

func GetHostAvailableResources() (ResourcePolling, error) {
	cpu, err := getSystemCPUPercent()
	if err != nil {
		return ResourcePolling{}, err
	}

	memory, err := getMemoryUsagePercent()
	if err != nil {
		return ResourcePolling{}, err
	}

	disk, err := getDiskUsagePercent("/")
	if err != nil {
		return ResourcePolling{}, err
	}

	return ResourcePolling{
		CPUUsagePercent:    cpu,
		MemoryUsagePercent: memory,
		DiskUsagePercent:   disk,
	}, nil
}

func IsRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}
