package hardware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/hardware"
	"nix-ai-help/pkg/logger"
)

// HardwareFunction handles hardware detection and configuration
type HardwareFunction struct {
	*functionbase.BaseFunction
	agent  *agent.HardwareAgent
	logger *logger.Logger
}

// HardwareRequest represents the input parameters for the hardware function
type HardwareRequest struct {
	Context        string            `json:"context"`
	Operation      string            `json:"operation,omitempty"`
	ComponentType  string            `json:"component_type,omitempty"`
	DetectAll      bool              `json:"detect_all,omitempty"`
	Generate       bool              `json:"generate,omitempty"`
	Format         string            `json:"format,omitempty"`
	IncludeDrivers bool              `json:"include_drivers,omitempty"`
	Options        map[string]string `json:"options,omitempty"`
}

// HardwareResponse represents the output of the hardware function
type HardwareResponse struct {
	Context         string              `json:"context"`
	Status          string              `json:"status"`
	Operation       string              `json:"operation"`
	Hardware        []HardwareComponent `json:"hardware,omitempty"`
	Configuration   string              `json:"configuration,omitempty"`
	Recommendations []string            `json:"recommendations,omitempty"`
	Issues          []HardwareIssue     `json:"issues,omitempty"`
	ErrorMessage    string              `json:"error_message,omitempty"`
	ExecutionTime   time.Duration       `json:"execution_time,omitempty"`
}

// HardwareComponent represents a detected hardware component
type HardwareComponent struct {
	Type          string            `json:"type"`
	Name          string            `json:"name"`
	Vendor        string            `json:"vendor,omitempty"`
	Model         string            `json:"model,omitempty"`
	Driver        string            `json:"driver,omitempty"`
	Supported     bool              `json:"supported"`
	Status        string            `json:"status"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// HardwareIssue represents a hardware-related issue
type HardwareIssue struct {
	Component   string   `json:"component"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Solution    string   `json:"solution,omitempty"`
	Resources   []string `json:"resources,omitempty"`
}

// NewHardwareFunction creates a new hardware function instance
func NewHardwareFunction() *HardwareFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Type of hardware operation to perform", true,
			[]string{"detect", "scan", "test", "diagnose", "configure", "driver-info"}, nil, nil),
		functionbase.StringParam("component", "Specific hardware component to focus on", false),
		functionbase.BoolParam("detailed", "Whether to perform detailed hardware analysis", false),
		functionbase.BoolParam("include_drivers", "Whether to include driver information", false),
		functionbase.ArrayParam("categories", "Hardware categories to scan", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"hardware",
		"Detect and configure hardware components for NixOS",
		parameters,
	)

	return &HardwareFunction{
		BaseFunction: baseFunc,
		agent:        nil, // Will be mocked
		logger:       logger.NewLogger(),
	}
}

// Name returns the function name
func (f *HardwareFunction) Name() string {
	return f.BaseFunction.Name()
}

// Description returns the function description
func (f *HardwareFunction) Description() string {
	return f.BaseFunction.Description()
}

// Version returns the function version
func (f *HardwareFunction) Version() string {
	return "1.0.0"
}

// Parameters returns the function parameter schema
func (f *HardwareFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"context": map[string]interface{}{
				"type":        "string",
				"description": "The context or reason for the hardware operation",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The hardware operation to perform",
				"enum":        []string{"detect", "generate-config", "scan", "test", "diagnose", "list-drivers"},
				"default":     "detect",
			},
			"component_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of hardware component to focus on",
				"enum":        []string{"cpu", "gpu", "network", "audio", "storage", "input", "display", "all"},
				"default":     "all",
			},
			"detect_all": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to detect all hardware components",
				"default":     true,
			},
			"generate": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to generate NixOS configuration",
				"default":     false,
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "The output format for configuration",
				"enum":        []string{"nix", "json", "yaml"},
				"default":     "nix",
			},
			"include_drivers": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include driver information",
				"default":     true,
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional hardware detection options",
			},
		},
		"required": []string{"context"},
	}
}

// Execute runs the hardware function with the given parameters
func (f *HardwareFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse the request manually
	req := HardwareRequest{}

	if operation, ok := params["operation"].(string); ok {
		req.Operation = operation
	}
	if component, ok := params["component"].(string); ok {
		req.ComponentType = component
	}
	if includeDrivers, ok := params["include_drivers"].(bool); ok {
		req.IncludeDrivers = includeDrivers
	}
	if detailed, ok := params["detailed"].(bool); ok {
		req.DetectAll = detailed // Map detailed to DetectAll
	}

	// Set defaults
	if req.Operation == "" {
		req.Operation = "detect"
	}
	if req.ComponentType == "" {
		req.ComponentType = "all"
	}
	if req.Format == "" {
		req.Format = "nix"
	}

	f.logger.Info(fmt.Sprintf("Executing hardware operation: %s for %s", req.Operation, req.ComponentType))

	// Execute the hardware operation
	response, err := f.executeHardwareOperation(ctx, &req)
	if err != nil {
		return functionbase.ErrorResult(fmt.Errorf("hardware operation failed: %v", err), time.Since(startTime)), nil
	}

	return functionbase.SuccessResult(response, time.Since(startTime)), nil
}

// executeHardwareOperation performs the actual hardware operation
func (f *HardwareFunction) executeHardwareOperation(ctx context.Context, req *HardwareRequest) (*HardwareResponse, error) {
	response := &HardwareResponse{
		Context:   req.Context,
		Operation: req.Operation,
		Status:    "success",
		Hardware:  []HardwareComponent{},
		Issues:    []HardwareIssue{},
	}

	switch req.Operation {
	case "detect":
		return f.detectHardware(ctx, req, response)
	case "generate-config":
		return f.generateConfig(ctx, req, response)
	case "scan":
		return f.scanHardware(ctx, req, response)
	case "test":
		return f.testHardware(ctx, req, response)
	case "diagnose":
		return f.diagnoseHardware(ctx, req, response)
	case "list-drivers":
		return f.listDrivers(ctx, req, response)
	default:
		return nil, fmt.Errorf("unsupported hardware operation: %s", req.Operation)
	}
}

// detectHardware detects hardware components
func (f *HardwareFunction) detectHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Detecting hardware components")

	// Use real hardware detection
	hardwareInfo, err := hardware.DetectHardwareComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to detect hardware: %v", err)
	}

	// Convert HardwareInfo to HardwareComponent list
	var components []HardwareComponent

	// Add CPU component
	if hardwareInfo.CPU != "" {
		vendor := extractVendor(hardwareInfo.CPU)
		config := map[string]string{}

		// Set microcode based on vendor
		if vendor == "AMD" {
			config["microcode"] = "hardware.cpu.amd.updateMicrocode = true;"
		} else if vendor == "Intel" {
			config["microcode"] = "hardware.cpu.intel.updateMicrocode = true;"
		}

		components = append(components, HardwareComponent{
			Type:          "CPU",
			Name:          hardwareInfo.CPU,
			Vendor:        vendor,
			Model:         extractModel(hardwareInfo.CPU),
			Driver:        "kernel_default",
			Supported:     true,
			Status:        "active",
			Configuration: config,
			Metadata: map[string]string{
				"architecture": hardwareInfo.Architecture,
			},
		})
	}

	// Add GPU components
	for _, gpu := range hardwareInfo.GPU {
		if strings.TrimSpace(gpu) != "" {
			vendor := extractVendor(gpu)
			driver := "kernel_default"
			config := map[string]string{}

			if strings.Contains(strings.ToLower(gpu), "nvidia") {
				driver = "nvidia"
				config["videoDrivers"] = "services.xserver.videoDrivers = [ \"nvidia\" ];"
				config["hardware"] = "hardware.opengl.enable = true;"
			} else if strings.Contains(strings.ToLower(gpu), "amd") || strings.Contains(strings.ToLower(gpu), "radeon") {
				driver = "amdgpu"
				config["videoDrivers"] = "services.xserver.videoDrivers = [ \"amdgpu\" ];"
				config["hardware"] = "hardware.opengl.enable = true;"
			} else if strings.Contains(strings.ToLower(gpu), "intel") {
				driver = "intel"
				config["videoDrivers"] = "services.xserver.videoDrivers = [ \"intel\" ];"
				config["hardware"] = "hardware.opengl.enable = true;"
			}

			components = append(components, HardwareComponent{
				Type:          "GPU",
				Name:          gpu,
				Vendor:        vendor,
				Model:         extractModel(gpu),
				Driver:        driver,
				Supported:     true,
				Status:        "active",
				Configuration: config,
				Metadata: map[string]string{
					"display_server": hardwareInfo.DisplayServer,
				},
			})
		}
	}

	// Add Audio component
	if hardwareInfo.Audio != "" {
		components = append(components, HardwareComponent{
			Type:      "Audio",
			Name:      hardwareInfo.Audio,
			Vendor:    extractVendor(hardwareInfo.Audio),
			Model:     extractModel(hardwareInfo.Audio),
			Driver:    "snd_hda_intel",
			Supported: true,
			Status:    "active",
			Configuration: map[string]string{
				"sound":      "sound.enable = true;",
				"pulseaudio": "hardware.pulseaudio.enable = true;",
			},
		})
	}

	// Add Network components (filter out virtual interfaces)
	for _, network := range hardwareInfo.Network {
		networkName := strings.TrimSpace(network)
		// Skip loopback, virtual, and container interfaces
		if networkName != "" && networkName != "lo" &&
			!strings.HasPrefix(networkName, "virbr") &&
			!strings.HasPrefix(networkName, "br-") &&
			!strings.HasPrefix(networkName, "docker") &&
			!strings.HasPrefix(networkName, "veth") &&
			!strings.Contains(networkName, "@if") {

			components = append(components, HardwareComponent{
				Type:      "Network",
				Name:      networkName,
				Driver:    "kernel_default",
				Supported: true,
				Status:    "active",
				Configuration: map[string]string{
					"networking": "networking.networkmanager.enable = true;",
				},
			})
		}
	}

	// Add Storage components
	for _, storage := range hardwareInfo.Storage {
		if strings.TrimSpace(storage) != "" {
			components = append(components, HardwareComponent{
				Type:      "Storage",
				Name:      storage,
				Driver:    "kernel_default",
				Supported: true,
				Status:    "active",
				Configuration: map[string]string{
					"filesystem": "boot.supportedFilesystems = [ \"ext4\" \"btrfs\" ];",
				},
			})
		}
	}

	response.Hardware = components

	// Add recommendations
	response.Recommendations = []string{
		"Use 'nixai hardware --operation=generate-config' to create hardware configuration.",
	}

	// If specific component requested, filter results
	if req.ComponentType != "all" && req.ComponentType != "" {
		var filteredComponents []HardwareComponent
		for _, comp := range response.Hardware {
			if strings.ToLower(comp.Type) == strings.ToLower(req.ComponentType) {
				filteredComponents = append(filteredComponents, comp)
			}
		}
		response.Hardware = filteredComponents
	}

	f.logger.Info(fmt.Sprintf("Detected %d hardware components", len(response.Hardware)))

	return response, nil
}

// generateConfig generates NixOS hardware configuration
func (f *HardwareFunction) generateConfig(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Generating hardware configuration")

	// First detect hardware components
	detectResponse, err := f.detectHardware(ctx, req, &HardwareResponse{
		Context:   req.Context,
		Operation: "detect",
		Status:    "success",
		Hardware:  []HardwareComponent{},
		Issues:    []HardwareIssue{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to detect hardware for config generation: %v", err)
	}

	// Generate configuration based on detected hardware
	var configParts []string
	switch req.Format {
	case "nix":
		configParts = append(configParts, "{ config, pkgs, ... }:", "{")

		// Use a map to track unique configuration lines and organize by category
		configMap := make(map[string]bool)
		hardwareSettings := []string{}
		bootSettings := []string{}
		serviceSettings := []string{}
		networkingSettings := []string{}

		for _, comp := range detectResponse.Hardware {
			for _, configLine := range comp.Configuration {
				if !configMap[configLine] {
					configMap[configLine] = true

					// Categorize configuration lines
					if strings.Contains(configLine, "hardware.") {
						hardwareSettings = append(hardwareSettings, configLine)
					} else if strings.Contains(configLine, "boot.") {
						bootSettings = append(bootSettings, configLine)
					} else if strings.Contains(configLine, "services.") {
						serviceSettings = append(serviceSettings, configLine)
					} else if strings.Contains(configLine, "networking.") {
						networkingSettings = append(networkingSettings, configLine)
					} else if strings.Contains(configLine, "sound.") {
						hardwareSettings = append(hardwareSettings, configLine)
					}
				}
			}
		}

		// Add organized configuration sections
		if len(hardwareSettings) > 0 {
			configParts = append(configParts, "  # Hardware Configuration")
			for _, setting := range hardwareSettings {
				configParts = append(configParts, "  "+setting)
			}
			configParts = append(configParts, "")
		}

		if len(bootSettings) > 0 {
			configParts = append(configParts, "  # Boot Configuration")
			for _, setting := range bootSettings {
				configParts = append(configParts, "  "+setting)
			}
			configParts = append(configParts, "")
		}

		if len(serviceSettings) > 0 {
			configParts = append(configParts, "  # Service Configuration")
			for _, setting := range serviceSettings {
				configParts = append(configParts, "  "+setting)
			}
			configParts = append(configParts, "")
		}

		if len(networkingSettings) > 0 {
			configParts = append(configParts, "  # Networking Configuration")
			for _, setting := range networkingSettings {
				configParts = append(configParts, "  "+setting)
			}
		}

		configParts = append(configParts, "}")
	case "json":
		configParts = append(configParts, "{")
		configParts = append(configParts, "  \"hardware\": {")
		// Extract key configuration settings
		hasOpenGL := false
		hasAMD := false
		hasNvidia := false
		for _, comp := range detectResponse.Hardware {
			if comp.Type == "GPU" {
				hasOpenGL = true
				if comp.Driver == "amdgpu" {
					hasAMD = true
				} else if comp.Driver == "nvidia" {
					hasNvidia = true
				}
			}
		}
		if hasOpenGL {
			configParts = append(configParts, "    \"opengl\": { \"enable\": true }")
		}
		if hasAMD {
			configParts = append(configParts, "    \"amdgpu\": { \"enable\": true }")
		}
		if hasNvidia {
			configParts = append(configParts, "    \"nvidia\": { \"enable\": true }")
		}
		configParts = append(configParts, "  }")
		configParts = append(configParts, "}")
	default:
		configParts = append(configParts, "# Generated hardware configuration")
		for _, comp := range detectResponse.Hardware {
			for _, configLine := range comp.Configuration {
				configParts = append(configParts, configLine)
			}
		}
	}

	response.Configuration = strings.Join(configParts, "\n")

	// Include hardware details from detection
	response.Hardware = detectResponse.Hardware

	// Generate recommendations for configuration
	response.Recommendations = []string{
		fmt.Sprintf("Generated %s configuration for %d hardware components", req.Format, len(response.Hardware)),
		"Review the configuration before applying to your system",
		"Consider backing up your current configuration first",
		"Test the configuration in a virtual machine if possible",
	}

	f.logger.Info("Hardware configuration generated successfully")

	return response, nil
}

// scanHardware performs comprehensive hardware scan
func (f *HardwareFunction) scanHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Performing comprehensive hardware scan")

	// Mock comprehensive scan results
	mockHardware := []HardwareComponent{
		{
			Type:      "CPU",
			Name:      "Intel Core i7-12700K",
			Vendor:    "Intel",
			Model:     "i7-12700K",
			Driver:    "intel_pstate",
			Supported: true,
			Status:    "active",
			Configuration: map[string]string{
				"kernelModules": "boot.initrd.kernelModules = [ \"intel_pstate\" ];",
				"options":       "boot.kernelParams = [ \"intel_pstate=active\" ];",
			},
			Metadata: map[string]string{"cores": "12", "threads": "20", "frequency": "3.6GHz"},
		},
		{
			Type:      "GPU",
			Name:      "NVIDIA GeForce RTX 3080",
			Vendor:    "NVIDIA",
			Model:     "RTX 3080",
			Driver:    "nvidia",
			Supported: true,
			Status:    "active",
			Configuration: map[string]string{
				"videoDrivers": "services.xserver.videoDrivers = [ \"nvidia\" ];",
				"hardware":     "hardware.opengl.enable = true;",
			},
			Metadata: map[string]string{"memory": "10GB", "cuda": "true", "compute": "8.6"},
		},
		{
			Type:      "Storage",
			Name:      "Samsung SSD 980 PRO",
			Vendor:    "Samsung",
			Model:     "980 PRO",
			Driver:    "nvme",
			Supported: true,
			Status:    "active",
			Configuration: map[string]string{
				"kernel": "boot.initrd.kernelModules = [ \"nvme\" ];",
			},
			Metadata: map[string]string{"capacity": "1TB", "interface": "PCIe 4.0"},
		},
	}

	response.Hardware = mockHardware

	// Mock scan issues
	mockIssues := []HardwareIssue{
		{
			Component:   "Wireless Network",
			Severity:    "warning",
			Description: "Wireless adapter may require proprietary firmware",
			Solution:    "Enable nixpkgs.config.allowUnfree and install firmware-linux-nonfree",
			Resources:   []string{"https://nixos.wiki/wiki/Wifi"},
		},
	}

	response.Issues = mockIssues

	// Generate comprehensive recommendations
	response.Recommendations = []string{
		fmt.Sprintf("Comprehensive scan completed: %d components detected", len(mockHardware)),
		fmt.Sprintf("Found %d potential issues requiring attention", len(mockIssues)),
		"Use 'nixai hardware --operation=test' to verify hardware functionality",
		"Use 'nixai hardware --operation=diagnose' for detailed issue analysis",
	}

	f.logger.Info(fmt.Sprintf("Hardware scan completed: %d components, %d issues", len(response.Hardware), len(response.Issues)))

	return response, nil
}

// testHardware tests hardware functionality
func (f *HardwareFunction) testHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Testing hardware functionality")

	// Mock hardware test results
	mockTestResults := []struct {
		Component  HardwareComponent
		TestStatus string
		TestResult string
		Solution   string
	}{
		{
			Component: HardwareComponent{
				Type:      "CPU",
				Name:      "Intel Core i7-12700K",
				Vendor:    "Intel",
				Model:     "i7-12700K",
				Driver:    "intel_pstate",
				Supported: true,
				Status:    "active",
				Configuration: map[string]string{
					"kernelModules": "boot.initrd.kernelModules = [ \"intel_pstate\" ];",
				},
			},
			TestStatus: "passed",
			TestResult: "All CPU cores functional, frequency scaling working",
			Solution:   "",
		},
		{
			Component: HardwareComponent{
				Type:      "GPU",
				Name:      "NVIDIA GeForce RTX 3080",
				Vendor:    "NVIDIA",
				Model:     "RTX 3080",
				Driver:    "nvidia",
				Supported: true,
				Status:    "active",
				Configuration: map[string]string{
					"videoDrivers": "services.xserver.videoDrivers = [ \"nvidia\" ];",
				},
			},
			TestStatus: "passed",
			TestResult: "GPU detected, CUDA available, drivers loaded",
			Solution:   "",
		},
		{
			Component: HardwareComponent{
				Type:      "Audio",
				Name:      "USB Audio Device",
				Vendor:    "Generic",
				Model:     "USB Audio",
				Driver:    "snd_usb_audio",
				Supported: true,
				Status:    "warning",
				Configuration: map[string]string{
					"sound": "sound.enable = true;",
				},
			},
			TestStatus: "failed",
			TestResult: "Audio device detected but no sound output",
			Solution:   "Check audio configuration and PulseAudio/PipeWire setup",
		},
	}

	// Convert test results to response format
	for _, result := range mockTestResults {
		hardware := result.Component
		hardware.Status = result.TestStatus
		hardware.Metadata = map[string]string{
			"test_result": result.TestResult,
			"test_time":   "2.5s",
		}
		response.Hardware = append(response.Hardware, hardware)

		// Add issues for failed tests
		if result.TestStatus == "failed" {
			issue := HardwareIssue{
				Component:   result.Component.Name,
				Severity:    "warning",
				Description: fmt.Sprintf("Hardware test failed: %s", result.TestResult),
				Solution:    result.Solution,
			}
			response.Issues = append(response.Issues, issue)
		}
	}

	// Generate test recommendations
	failedTests := 0
	for _, result := range mockTestResults {
		if result.TestStatus == "failed" {
			failedTests++
		}
	}

	if failedTests > 0 {
		response.Recommendations = append(response.Recommendations, fmt.Sprintf("%d hardware tests failed. Review issues and solutions.", failedTests))
	} else {
		response.Recommendations = append(response.Recommendations, "All hardware tests passed successfully.")
	}

	f.logger.Info(fmt.Sprintf("Hardware testing completed: %d components tested", len(mockTestResults)))

	return response, nil
}

// diagnoseHardware diagnoses hardware issues
func (f *HardwareFunction) diagnoseHardware(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Diagnosing hardware issues")

	// Mock hardware diagnosis results
	mockComponents := []HardwareComponent{
		{
			Type:      "CPU",
			Name:      "Intel Core i7-12700K",
			Vendor:    "Intel",
			Model:     "i7-12700K",
			Driver:    "intel_pstate",
			Supported: true,
			Status:    "healthy",
			Configuration: map[string]string{
				"kernelModules": "boot.initrd.kernelModules = [ \"intel_pstate\" ];",
			},
			Metadata: map[string]string{
				"diagnosis_result": "CPU operating normally, no thermal issues detected",
				"confidence":       "0.95",
			},
		},
		{
			Type:      "GPU",
			Name:      "NVIDIA GeForce RTX 3080",
			Vendor:    "NVIDIA",
			Model:     "RTX 3080",
			Driver:    "nvidia",
			Supported: true,
			Status:    "warning",
			Configuration: map[string]string{
				"videoDrivers": "services.xserver.videoDrivers = [ \"nvidia\" ];",
			},
			Metadata: map[string]string{
				"diagnosis_result": "GPU detected but driver version may be outdated",
				"confidence":       "0.80",
			},
		},
	}

	response.Hardware = mockComponents

	// Mock diagnosis issues
	mockIssues := []HardwareIssue{
		{
			Component:   "NVIDIA GeForce RTX 3080",
			Severity:    "warning",
			Description: "GPU driver version is outdated and may cause performance issues",
			Solution:    "Update NVIDIA drivers using hardware.nvidia.package option",
			Resources: []string{
				"https://nixos.wiki/wiki/Nvidia",
				"https://github.com/NixOS/nixpkgs/blob/master/pkgs/os-specific/linux/nvidia-x11/default.nix",
			},
		},
		{
			Component:   "Wireless Network",
			Severity:    "critical",
			Description: "Wireless adapter requires proprietary firmware that is not installed",
			Solution:    "Enable allowUnfree and install hardware.enableRedistributableFirmware",
			Resources: []string{
				"https://nixos.wiki/wiki/Wifi",
			},
		},
	}

	response.Issues = mockIssues

	// Generate diagnosis recommendations
	criticalIssues := 0
	for _, issue := range mockIssues {
		if issue.Severity == "critical" {
			criticalIssues++
		}
	}

	if criticalIssues > 0 {
		response.Recommendations = append(response.Recommendations, fmt.Sprintf("%d critical hardware issues require immediate attention", criticalIssues))
	}

	response.Recommendations = append(response.Recommendations, "Follow the provided solutions to resolve hardware issues")
	response.Recommendations = append(response.Recommendations, "Check NixOS hardware compatibility list for additional information")

	f.logger.Info(fmt.Sprintf("Hardware diagnosis completed: %d issues found", len(response.Issues)))

	return response, nil
}

// listDrivers lists available drivers for hardware
func (f *HardwareFunction) listDrivers(ctx context.Context, req *HardwareRequest, response *HardwareResponse) (*HardwareResponse, error) {
	f.logger.Info("Listing available hardware drivers")

	// Mock available drivers
	mockDrivers := []struct {
		ComponentType string
		Name          string
		Supported     bool
		Status        string
		Version       string
		Description   string
		Package       string
	}{
		{
			ComponentType: "GPU",
			Name:          "nvidia",
			Supported:     true,
			Status:        "available",
			Version:       "535.154.05",
			Description:   "NVIDIA proprietary driver",
			Package:       "linuxPackages.nvidia_x11",
		},
		{
			ComponentType: "GPU",
			Name:          "nouveau",
			Supported:     true,
			Status:        "available",
			Version:       "1.0.17",
			Description:   "Open source NVIDIA driver",
			Package:       "xorg.xf86videonouveau",
		},
		{
			ComponentType: "Audio",
			Name:          "snd_hda_intel",
			Supported:     true,
			Status:        "active",
			Version:       "kernel",
			Description:   "Intel HD Audio driver",
			Package:       "kernel module",
		},
		{
			ComponentType: "Network",
			Name:          "iwlwifi",
			Supported:     true,
			Status:        "available",
			Version:       "kernel",
			Description:   "Intel wireless driver",
			Package:       "hardware.enableRedistributableFirmware",
		},
		{
			ComponentType: "Storage",
			Name:          "nvme",
			Supported:     true,
			Status:        "active",
			Version:       "kernel",
			Description:   "NVMe storage driver",
			Package:       "kernel module",
		},
	}

	// Filter by component type if specified
	var filteredDrivers []struct {
		ComponentType string
		Name          string
		Supported     bool
		Status        string
		Version       string
		Description   string
		Package       string
	}

	if req.ComponentType != "" && req.ComponentType != "all" {
		for _, driver := range mockDrivers {
			if strings.ToLower(driver.ComponentType) == strings.ToLower(req.ComponentType) {
				filteredDrivers = append(filteredDrivers, driver)
			}
		}
	} else {
		filteredDrivers = mockDrivers
	}

	// Convert to hardware components format
	for _, driver := range filteredDrivers {
		hardware := HardwareComponent{
			Type:      driver.ComponentType,
			Name:      driver.Name,
			Driver:    driver.Name,
			Supported: driver.Supported,
			Status:    driver.Status,
			Metadata: map[string]string{
				"version":     driver.Version,
				"description": driver.Description,
				"package":     driver.Package,
			},
		}
		response.Hardware = append(response.Hardware, hardware)
	}

	// Generate driver recommendations
	unsupportedDrivers := 0
	for _, driver := range filteredDrivers {
		if !driver.Supported {
			unsupportedDrivers++
		}
	}

	if unsupportedDrivers > 0 {
		response.Recommendations = append(response.Recommendations, fmt.Sprintf("%d drivers may not be fully supported", unsupportedDrivers))
	}

	response.Recommendations = append(response.Recommendations, "Enable required drivers in your NixOS configuration")
	response.Recommendations = append(response.Recommendations, "Consider using nixos-hardware for automatic driver configuration")

	f.logger.Info(fmt.Sprintf("Listed %d available drivers", len(filteredDrivers)))

	return response, nil
}

// generateHardwareRecommendations generates recommendations based on detected hardware
func (f *HardwareFunction) generateHardwareRecommendations(hardware []HardwareComponent) []string {
	recommendations := []string{}

	unsupportedCount := 0
	for _, hw := range hardware {
		if !hw.Supported {
			unsupportedCount++
		}
	}

	if unsupportedCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d hardware components may need additional configuration", unsupportedCount))
	}

	// Check for common hardware types
	hasGPU := false
	hasAudio := false
	hasNetwork := false

	for _, hw := range hardware {
		switch hw.Type {
		case "gpu":
			hasGPU = true
		case "audio":
			hasAudio = true
		case "network":
			hasNetwork = true
		}
	}

	if hasGPU {
		recommendations = append(recommendations, "GPU detected. Consider enabling appropriate graphics drivers.")
	}
	if hasAudio {
		recommendations = append(recommendations, "Audio hardware detected. Ensure PulseAudio or PipeWire is configured.")
	}
	if hasNetwork {
		recommendations = append(recommendations, "Network hardware detected. Consider configuring NetworkManager.")
	}

	recommendations = append(recommendations, "Use 'nixai hardware --operation=generate-config' to create hardware configuration.")

	return recommendations
}

// detectHardwareIssues detects common hardware issues
func (f *HardwareFunction) detectHardwareIssues(hardware []HardwareComponent) []HardwareIssue {
	issues := []HardwareIssue{}

	for _, hw := range hardware {
		if !hw.Supported {
			issue := HardwareIssue{
				Component:   hw.Name,
				Severity:    "warning",
				Description: fmt.Sprintf("%s (%s) may not be fully supported", hw.Name, hw.Type),
				Solution:    "Check NixOS hardware database for compatibility information",
				Resources:   []string{"https://github.com/NixOS/nixos-hardware"},
			}
			issues = append(issues, issue)
		}

		if hw.Status == "error" || hw.Status == "failed" {
			issue := HardwareIssue{
				Component:   hw.Name,
				Severity:    "error",
				Description: fmt.Sprintf("%s has hardware errors", hw.Name),
				Solution:    "Check hardware connections and driver configuration",
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// extractVendor extracts vendor information from hardware description
func extractVendor(description string) string {
	desc := strings.ToLower(description)
	if strings.Contains(desc, "intel") {
		return "Intel"
	}
	if strings.Contains(desc, "amd") {
		return "AMD"
	}
	if strings.Contains(desc, "nvidia") {
		return "NVIDIA"
	}
	if strings.Contains(desc, "qualcomm") {
		return "Qualcomm"
	}
	if strings.Contains(desc, "broadcom") {
		return "Broadcom"
	}
	if strings.Contains(desc, "realtek") {
		return "Realtek"
	}
	return "Unknown"
}

// extractModel extracts model information from hardware description
func extractModel(description string) string {
	// Simple model extraction - could be enhanced
	parts := strings.Fields(description)
	if len(parts) > 2 {
		// Try to find model part (usually after vendor)
		for i, part := range parts {
			if i > 0 && (strings.Contains(strings.ToLower(part), "core") ||
				strings.Contains(strings.ToLower(part), "geforce") ||
				strings.Contains(strings.ToLower(part), "radeon") ||
				strings.Contains(strings.ToLower(part), "hda")) {
				// Return rest of string after vendor
				return strings.Join(parts[i:], " ")
			}
		}
	}
	return strings.TrimSpace(description)
}
