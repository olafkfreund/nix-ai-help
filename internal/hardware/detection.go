// Package hardware provides hardware detection and configuration utilities
package hardware

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HardwareInfo represents detected hardware information
type HardwareInfo struct {
	CPU            string
	GPU            []string
	Memory         string
	Storage        []string
	Network        []string
	Audio          string
	USB            []string
	PCI            []string
	Firmware       string
	DisplayServer  string // X11 or Wayland
	Architecture   string
	Virtualization string // VM detection and virtualization capabilities
}

// DetectHardwareComponents performs comprehensive hardware detection
func DetectHardwareComponents() (*HardwareInfo, error) {
	info := &HardwareInfo{}

	// Detect CPU
	if cpu, err := runCommand("lscpu | grep 'Model name' | cut -d':' -f2 | xargs"); err == nil {
		info.CPU = cpu
	}

	// Detect GPU devices with more detailed information
	if gpu, err := runCommand("lspci | grep -i vga"); err == nil && gpu != "" {
		info.GPU = append(info.GPU, gpu)
	}
	if gpu3d, err := runCommand("lspci | grep -i '3d'"); err == nil && gpu3d != "" {
		info.GPU = append(info.GPU, gpu3d)
	}
	// Detect additional display controllers
	if display, err := runCommand("lspci | grep -i 'display\\|graphics'"); err == nil && display != "" {
		// Add only if not already in GPU list
		displayLines := strings.Split(display, "\n")
		for _, line := range displayLines {
			line = strings.TrimSpace(line)
			if line != "" && !sliceContains(info.GPU, line) {
				info.GPU = append(info.GPU, line)
			}
		}
	}

	// Detect memory with more details
	if mem, err := runCommand("free -h | head -2 | tail -1 | awk '{print $2}'"); err == nil {
		// Get additional memory info
		if memType, err := runCommand("dmidecode -t memory | grep -i 'type:' | head -1 | cut -d':' -f2 | xargs"); err == nil && memType != "" {
			info.Memory = fmt.Sprintf("%s (%s)", mem, memType)
		} else {
			info.Memory = mem
		}
	}

	// Detect storage
	if storage, err := runCommand("lsblk -d -o name,size,type | grep disk"); err == nil {
		info.Storage = strings.Split(storage, "\n")
	}

	// Detect network interfaces
	if network, err := runCommand("ip link show | grep -E '^[0-9]+:' | cut -d':' -f2 | xargs"); err == nil {
		info.Network = strings.Fields(network)
	}

	// Detect audio
	if audio, err := runCommand("lspci | grep -i audio"); err == nil {
		info.Audio = audio
	}

	// Detect USB devices
	if usb, err := runCommand("lsusb"); err == nil {
		info.USB = strings.Split(usb, "\n")
	}

	// Detect PCI devices
	if pci, err := runCommand("lspci"); err == nil {
		info.PCI = strings.Split(pci, "\n")
	}

	// Detect firmware type
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		info.Firmware = "UEFI"
	} else {
		info.Firmware = "BIOS"
	}

	// Detect display server
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		info.DisplayServer = "Wayland"
	} else if os.Getenv("DISPLAY") != "" {
		info.DisplayServer = "X11"
	} else {
		info.DisplayServer = "Unknown/Console"
	}

	// Detect architecture
	if arch, err := runCommand("uname -m"); err == nil {
		info.Architecture = strings.TrimSpace(arch)
	}

	// Detect virtualization
	var virtInfo []string

	// Check if running in a VM
	if virt, err := runCommand("systemd-detect-virt 2>/dev/null"); err == nil && virt != "none" {
		virtInfo = append(virtInfo, fmt.Sprintf("Running in: %s", virt))
	}

	// Check CPU virtualization support
	if cpuVirt, err := runCommand("lscpu | grep -E 'Virtualization|VT-x|AMD-V'"); err == nil && cpuVirt != "" {
		virtInfo = append(virtInfo, fmt.Sprintf("CPU Features: %s", cpuVirt))
	}

	// Check for hypervisor
	if hyper, err := runCommand("lscpu | grep 'Hypervisor vendor'"); err == nil && hyper != "" {
		virtInfo = append(virtInfo, hyper)
	}

	if len(virtInfo) > 0 {
		info.Virtualization = strings.Join(virtInfo, "\n")
	} else {
		info.Virtualization = "Native/Unknown"
	}

	return info, nil
}

// runCommand executes a shell command and returns its output
func runCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// sliceContains checks if a slice contains a string
func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
