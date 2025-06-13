package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestHardwareAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "hardware agent response"}
	agent := NewHardwareAgent(mockProvider)

	hwCtx := &HardwareContext{
		SystemInfo:    "Dell XPS 13 9310",
		CPUInfo:       "Intel Core i7-1165G7 @ 2.80GHz",
		GPUInfo:       "Intel Iris Xe Graphics",
		MemoryInfo:    "16GB LPDDR4X-4267",
		StorageInfo:   "512GB NVMe SSD",
		NetworkInfo:   "Intel Wi-Fi 6 AX201, Realtek Ethernet",
		AudioInfo:     "Realtek ALC3254",
		USBDevices:    []string{"Logitech USB Mouse", "USB-C Hub"},
		PCIDevices:    []string{"Intel Corporation Tiger Lake-LP Thunderbolt 4 Controller"},
		KernelModules: []string{"i915", "iwlwifi", "rtl8xxxu"},
		LoadedDrivers: []string{"intel_graphics", "iwlwifi", "snd_hda_intel"},
		Architecture:  "x86_64-linux",
		SecureBoot:    true,
		BIOS_UEFI:     "UEFI 2.7",
	}
	agent.SetContext(hwCtx)

	input := "How can I optimize my graphics performance on this laptop?"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "hardware agent")
}

func TestHardwareAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "hardware agent response"}
	agent := NewHardwareAgent(mockProvider)

	hwCtx := &HardwareContext{
		SystemInfo:          "Custom Gaming PC",
		CPUInfo:             "AMD Ryzen 7 5800X3D",
		GPUInfo:             "NVIDIA RTX 4070 Ti",
		MemoryInfo:          "32GB DDR4-3600",
		StorageInfo:         "1TB NVMe SSD + 2TB HDD",
		MissingDrivers:      []string{"nvidia-drivers"},
		HardwareIssues:      []string{"GPU not detected properly"},
		OptimizationGoals:   []string{"gaming performance", "low latency"},
		RecommendedPackages: []string{"nvidia-drivers", "steam", "gamemode"},
		ConfigOptions:       map[string]string{"services.xserver.videoDrivers": "[\"nvidia\"]"},
	}
	agent.SetContext(hwCtx)

	input := "Help me set up NVIDIA drivers for gaming on NixOS"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "hardware agent response")
}

func TestHardwareAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewHardwareAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RoleHardware)
	require.NoError(t, err)
	require.Equal(t, roles.RoleHardware, agent.role)

	// Test setting context
	hwCtx := &HardwareContext{CPUInfo: "Intel i5"}
	agent.SetContext(hwCtx)
	require.Equal(t, hwCtx, agent.contextData)
}

func TestHardwareAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewHardwareAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestHardwareContext_Formatting(t *testing.T) {
	hwCtx := &HardwareContext{
		SystemInfo:          "ThinkPad X1 Carbon Gen 9",
		CPUInfo:             "Intel Core i7-1185G7 @ 3.00GHz",
		GPUInfo:             "Intel Iris Xe Graphics",
		MemoryInfo:          "32GB LPDDR4X-4266",
		StorageInfo:         "1TB PCIe NVMe SSD",
		NetworkInfo:         "Intel Wi-Fi 6E AX210, Intel Ethernet I219-LM",
		AudioInfo:           "Realtek ALC3287",
		USBDevices:          []string{"Lenovo USB-C Dock", "Wireless Mouse", "USB Webcam"},
		PCIDevices:          []string{"Intel Tiger Lake-UP3 Thunderbolt 4 Controller", "Intel Wi-Fi 6E AX210"},
		KernelModules:       []string{"i915", "iwlwifi", "e1000e", "snd_hda_intel", "thinkpad_acpi"},
		LoadedDrivers:       []string{"intel_graphics", "iwlwifi", "ethernet", "audio"},
		MissingDrivers:      []string{},
		HardwareIssues:      []string{"Occasional Wi-Fi disconnects"},
		PowerManagement:     "TLP enabled, CPU scaling active",
		ThermalInfo:         "CPU thermal throttling at 85°C",
		BIOS_UEFI:           "UEFI 2.8, TPM 2.0 enabled",
		SecureBoot:          true,
		VirtualizationInfo:  "Intel VT-x, VT-d enabled",
		Architecture:        "x86_64-linux",
		Microcode:           "0x34",
		Firmware:            []string{"firmware-linux-nonfree", "intel-microcode"},
		HardwareConfig:      map[string]string{"hardware.cpu.intel.updateMicrocode": "true", "hardware.enableRedistributableFirmware": "true"},
		OptimizationGoals:   []string{"battery life", "thermal management", "performance"},
		CompatibilityIssues: []string{"Fingerprint reader needs proprietary driver"},
		RecommendedPackages: []string{"tlp", "powertop", "thermald", "intel-gpu-tools"},
		ConfigOptions:       map[string]string{"services.tlp.enable": "true", "services.thermald.enable": "true"},
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, hwCtx.SystemInfo)
	require.Contains(t, hwCtx.CPUInfo, "Intel Core i7")
	require.Contains(t, hwCtx.GPUInfo, "Intel Iris Xe")
	require.Contains(t, hwCtx.MemoryInfo, "32GB")
	require.Contains(t, hwCtx.StorageInfo, "1TB")
	require.Contains(t, hwCtx.NetworkInfo, "Wi-Fi 6E")
	require.Contains(t, hwCtx.AudioInfo, "Realtek")
	require.Len(t, hwCtx.USBDevices, 3)
	require.Len(t, hwCtx.PCIDevices, 2)
	require.Len(t, hwCtx.KernelModules, 5)
	require.Len(t, hwCtx.LoadedDrivers, 4)
	require.Empty(t, hwCtx.MissingDrivers)
	require.Len(t, hwCtx.HardwareIssues, 1)
	require.Contains(t, hwCtx.PowerManagement, "TLP")
	require.Contains(t, hwCtx.ThermalInfo, "85°C")
	require.Contains(t, hwCtx.BIOS_UEFI, "UEFI")
	require.True(t, hwCtx.SecureBoot)
	require.Contains(t, hwCtx.VirtualizationInfo, "VT-x")
	require.Equal(t, "x86_64-linux", hwCtx.Architecture)
	require.Equal(t, "0x34", hwCtx.Microcode)
	require.Len(t, hwCtx.Firmware, 2)
	require.Len(t, hwCtx.HardwareConfig, 2)
	require.Len(t, hwCtx.OptimizationGoals, 3)
	require.Len(t, hwCtx.CompatibilityIssues, 1)
	require.Len(t, hwCtx.RecommendedPackages, 4)
	require.Len(t, hwCtx.ConfigOptions, 2)
}
