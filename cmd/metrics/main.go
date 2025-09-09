package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// GetNVIDIAGPUUsage returns the current NVIDIA GPU utilization as a percentage (0-100)
func GetNVIDIAGPUUsage() (float64, error) {
	// Use nvidia-smi to get GPU utilization
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run nvidia-smi: %w", err)
	}

	// Parse the output
	utilStr := strings.TrimSpace(string(output))
	utilization, err := strconv.ParseFloat(utilStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse GPU utilization: %w", err)
	}

	return utilization, nil
}

// GetAMDGPUUsage returns the current AMD GPU utilization as a percentage (0-100)
func GetAMDGPUUsage() (float64, error) {
	// Try rocm-smi first (for newer AMD cards)
	cmd := exec.Command("rocm-smi", "--showuse", "--csv")
	output, err := cmd.Output()
	if err == nil {
		// Parse rocm-smi output
		lines := strings.SplitSeq(string(output), "\n")
		for line := range lines {
			if strings.Contains(line, "GPU use (%)") {
				fields := strings.Split(line, ",")
				if len(fields) >= 2 {
					utilStr := strings.TrimSpace(fields[1])
					utilStr = strings.TrimSuffix(utilStr, "%")
					if utilization, err := strconv.ParseFloat(utilStr, 64); err == nil {
						return utilization, nil
					}
				}
			}
		}
	}

	// Fallback to radeontop (requires radeontop to be installed)
	cmd = exec.Command("timeout", "2", "radeontop", "-d", "-", "-l", "1")
	output, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get AMD GPU usage (tried rocm-smi and radeontop): %w", err)
	}

	// Parse radeontop output - look for "gpu" usage line
	re := regexp.MustCompile(`gpu\s+(\d+(?:\.\d+)?)%`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse AMD GPU utilization from radeontop output")
	}

	utilization, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse AMD GPU utilization: %w", err)
	}

	return utilization, nil
}

// GetIntelGPUUsage returns the current Intel GPU utilization as a percentage (0-100)
func GetIntelGPUUsage() (float64, error) {
	// Try intel_gpu_top first (part of igt-gpu-tools)
	cmd := exec.Command("timeout", "2", "intel_gpu_top", "-s", "1000", "-o", "-")
	output, err := cmd.Output()
	if err == nil {
		// Parse intel_gpu_top output
		lines := strings.SplitSeq(string(output), "\n")
		for line := range lines {
			if strings.Contains(line, "Render/3D") {
				// Look for percentage in the line
				re := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					utilization, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						return utilization, nil
					}
				}
			}
		}
	}

	// Fallback to reading from sysfs (Linux)
	// Note: This path may vary depending on the system
	cmd = exec.Command("cat", "/sys/class/drm/card0/gt/gt0/rps_cur_freq_mhz")
	curFreq, err1 := cmd.Output()

	cmd = exec.Command("cat", "/sys/class/drm/card0/gt/gt0/rps_max_freq_mhz")
	maxFreq, err2 := cmd.Output()

	if err1 == nil && err2 == nil {
		curFreqVal, err1 := strconv.ParseFloat(strings.TrimSpace(string(curFreq)), 64)
		maxFreqVal, err2 := strconv.ParseFloat(strings.TrimSpace(string(maxFreq)), 64)

		if err1 == nil && err2 == nil && maxFreqVal > 0 {
			// Approximate usage based on frequency ratio
			usage := (curFreqVal / maxFreqVal) * 100
			return usage, nil
		}
	}

	return 0, fmt.Errorf("failed to get Intel GPU usage (tried intel_gpu_top and sysfs)")
}

// Example usage
func main() {
	// Test NVIDIA GPU usage
	if usage, err := GetNVIDIAGPUUsage(); err == nil {
		fmt.Printf("NVIDIA GPU Usage: %.2f%%\n", usage)
	} else {
		fmt.Printf("NVIDIA GPU Error: %v\n", err)
	}

	// Test AMD GPU usage
	if usage, err := GetAMDGPUUsage(); err == nil {
		fmt.Printf("AMD GPU Usage: %.2f%%\n", usage)
	} else {
		fmt.Printf("AMD GPU Error: %v\n", err)
	}

	// Test Intel GPU usage
	if usage, err := GetIntelGPUUsage(); err == nil {
		fmt.Printf("Intel GPU Usage: %.2f%%\n", usage)
	} else {
		fmt.Printf("Intel GPU Error: %v\n", err)
	}
}
