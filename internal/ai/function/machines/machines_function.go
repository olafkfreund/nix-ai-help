package machines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// MachinesFunction handles machine management and configuration operations
type MachinesFunction struct {
	*functionbase.BaseFunction
	machinesAgent *agent.MachinesAgent
	logger        *logger.Logger
}

// MachinesRequest represents the input parameters for the machines function
type MachinesRequest struct {
	Operation     string            `json:"operation"`
	MachineName   string            `json:"machine_name,omitempty"`
	MachineType   string            `json:"machine_type,omitempty"`
	Architecture  string            `json:"architecture,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Template      string            `json:"template,omitempty"`
	Environment   string            `json:"environment,omitempty"`
	Services      []string          `json:"services,omitempty"`
	Hardware      map[string]string `json:"hardware,omitempty"`
	Network       map[string]string `json:"network,omitempty"`
	Security      map[string]string `json:"security,omitempty"`
	Performance   map[string]string `json:"performance,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
}

// MachinesResponse represents the output of the machines function
type MachinesResponse struct {
	Operation         string                 `json:"operation"`
	Status            string                 `json:"status"`
	Error             string                 `json:"error,omitempty"`
	Machines          []MachineInfo          `json:"machines,omitempty"`
	Configuration     *MachineConfiguration  `json:"configuration,omitempty"`
	Templates         []MachineTemplate      `json:"templates,omitempty"`
	Requirements      []string               `json:"requirements,omitempty"`
	Recommendations   []string               `json:"recommendations,omitempty"`
	SecuritySettings  []SecuritySetting      `json:"security_settings,omitempty"`
	PerformanceConfig []PerformanceSetting   `json:"performance_config,omitempty"`
	NetworkConfig     *NetworkConfiguration  `json:"network_config,omitempty"`
	HardwareInfo      *HardwareInfo          `json:"hardware_info,omitempty"`
	SetupSteps        []string               `json:"setup_steps,omitempty"`
	Commands          []string               `json:"commands,omitempty"`
	Documentation     []DocumentationLink    `json:"documentation,omitempty"`
	Examples          []ConfigurationExample `json:"examples,omitempty"`
}

// MachineInfo represents information about a machine
type MachineInfo struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Architecture  string            `json:"architecture"`
	Status        string            `json:"status"`
	Environment   string            `json:"environment"`
	Services      []string          `json:"services"`
	Configuration map[string]string `json:"configuration"`
	Location      string            `json:"location,omitempty"`
	LastUpdate    string            `json:"last_update,omitempty"`
}

// MachineConfiguration represents a machine configuration
type MachineConfiguration struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Architecture  string                 `json:"architecture"`
	Hardware      map[string]interface{} `json:"hardware"`
	Network       map[string]interface{} `json:"network"`
	Security      map[string]interface{} `json:"security"`
	Performance   map[string]interface{} `json:"performance"`
	Services      []string               `json:"services"`
	Environment   map[string]string      `json:"environment"`
	Configuration string                 `json:"configuration"`
	Files         map[string]string      `json:"files"`
}

// MachineTemplate represents a machine template
type MachineTemplate struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Type          string            `json:"type"`
	Architecture  string            `json:"architecture"`
	Category      string            `json:"category"`
	Tags          []string          `json:"tags"`
	Requirements  []string          `json:"requirements"`
	Features      []string          `json:"features"`
	UseCases      []string          `json:"use_cases"`
	Configuration string            `json:"configuration"`
	Variables     map[string]string `json:"variables"`
}

// SecuritySetting represents a security setting
type SecuritySetting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Required    bool   `json:"required"`
	Recommended bool   `json:"recommended"`
}

// PerformanceSetting represents a performance setting
type PerformanceSetting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Impact      string `json:"impact"`
	Recommended bool   `json:"recommended"`
}

// NetworkConfiguration represents network configuration
type NetworkConfiguration struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	Routing    []RouteConfig      `json:"routing"`
	Firewall   *FirewallConfig    `json:"firewall"`
	DNS        *DNSConfig         `json:"dns"`
	VPN        *VPNConfig         `json:"vpn,omitempty"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Address string            `json:"address"`
	Gateway string            `json:"gateway,omitempty"`
	DNS     []string          `json:"dns,omitempty"`
	Options map[string]string `json:"options,omitempty"`
	Enabled bool              `json:"enabled"`
}

// RouteConfig represents routing configuration
type RouteConfig struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric,omitempty"`
}

// FirewallConfig represents firewall configuration
type FirewallConfig struct {
	Enabled bool           `json:"enabled"`
	Rules   []FirewallRule `json:"rules"`
	Zones   []FirewallZone `json:"zones"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Name        string `json:"name"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	Port        string `json:"port,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// FirewallZone represents a firewall zone
type FirewallZone struct {
	Name       string   `json:"name"`
	Interfaces []string `json:"interfaces"`
	Services   []string `json:"services"`
	Ports      []string `json:"ports"`
}

// DNSConfig represents DNS configuration
type DNSConfig struct {
	Servers []string          `json:"servers"`
	Domains []string          `json:"domains"`
	Search  []string          `json:"search"`
	Options map[string]string `json:"options"`
	Enabled bool              `json:"enabled"`
}

// VPNConfig represents VPN configuration
type VPNConfig struct {
	Type     string            `json:"type"`
	Server   string            `json:"server"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
	Auth     map[string]string `json:"auth"`
	Routes   []string          `json:"routes"`
	DNS      []string          `json:"dns"`
	Enabled  bool              `json:"enabled"`
}

// HardwareInfo represents hardware information
type HardwareInfo struct {
	CPU       *CPUInfo        `json:"cpu"`
	Memory    *MemoryInfo     `json:"memory"`
	Storage   []StorageDevice `json:"storage"`
	Network   []NetworkDevice `json:"network"`
	Graphics  []GraphicsCard  `json:"graphics"`
	USB       []USBDevice     `json:"usb"`
	Audio     []AudioDevice   `json:"audio"`
	Bluetooth *BluetoothInfo  `json:"bluetooth,omitempty"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model        string   `json:"model"`
	Cores        int      `json:"cores"`
	Threads      int      `json:"threads"`
	Frequency    string   `json:"frequency"`
	Cache        string   `json:"cache"`
	Architecture string   `json:"architecture"`
	Features     []string `json:"features"`
	Temperature  string   `json:"temperature,omitempty"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total     string   `json:"total"`
	Available string   `json:"available"`
	Used      string   `json:"used"`
	Type      string   `json:"type"`
	Speed     string   `json:"speed"`
	Modules   []string `json:"modules"`
}

// StorageDevice represents a storage device
type StorageDevice struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Size        string `json:"size"`
	Interface   string `json:"interface"`
	Model       string `json:"model"`
	Serial      string `json:"serial,omitempty"`
	Health      string `json:"health,omitempty"`
	Temperature string `json:"temperature,omitempty"`
}

// NetworkDevice represents a network device
type NetworkDevice struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Driver   string   `json:"driver"`
	Speed    string   `json:"speed"`
	Status   string   `json:"status"`
	MAC      string   `json:"mac"`
	Features []string `json:"features"`
}

// GraphicsCard represents a graphics card
type GraphicsCard struct {
	Name        string `json:"name"`
	Driver      string `json:"driver"`
	Memory      string `json:"memory"`
	Core        string `json:"core"`
	Temperature string `json:"temperature,omitempty"`
	Power       string `json:"power,omitempty"`
}

// USBDevice represents a USB device
type USBDevice struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Class   string `json:"class"`
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Speed   string `json:"speed"`
	Port    string `json:"port"`
}

// AudioDevice represents an audio device
type AudioDevice struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Type       string `json:"type"`
	Channels   string `json:"channels"`
	SampleRate string `json:"sample_rate"`
}

// BluetoothInfo represents Bluetooth information
type BluetoothInfo struct {
	Enabled      bool     `json:"enabled"`
	Version      string   `json:"version"`
	Devices      []string `json:"devices"`
	Controller   string   `json:"controller"`
	Discoverable bool     `json:"discoverable"`
}

// DocumentationLink represents a documentation link
type DocumentationLink struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// ConfigurationExample represents a configuration example
type ConfigurationExample struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Code        string   `json:"code"`
	Language    string   `json:"language"`
	Tags        []string `json:"tags"`
}

// NewMachinesFunction creates a new machines function
func NewMachinesFunction() *MachinesFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParamWithOptions("operation", "Operation to perform", true,
			[]string{"list", "create", "configure", "template", "analyze", "optimize", "security", "performance", "network", "hardware"}, nil, nil),
		functionbase.StringParam("machine_name", "Name of the machine", false),
		functionbase.StringParamWithOptions("machine_type", "Type of machine", false,
			[]string{"desktop", "laptop", "server", "workstation", "vm", "container", "cloud", "embedded"}, nil, nil),
		functionbase.StringParamWithOptions("architecture", "System architecture", false,
			[]string{"x86_64", "aarch64", "i686", "armv7l", "riscv64"}, nil, nil),
		{
			Name:        "configuration",
			Type:        "object",
			Description: "Machine configuration settings",
			Required:    false,
		},
		functionbase.StringParam("template", "Template name for machine creation", false),
		functionbase.StringParamWithOptions("environment", "Target environment", false,
			[]string{"development", "testing", "staging", "production", "personal", "enterprise"}, nil, nil),
		{
			Name:        "services",
			Type:        "array",
			Description: "Services to configure",
			Required:    false,
		},
		{
			Name:        "hardware",
			Type:        "object",
			Description: "Hardware specifications and requirements",
			Required:    false,
		},
		{
			Name:        "network",
			Type:        "object",
			Description: "Network configuration settings",
			Required:    false,
		},
		{
			Name:        "security",
			Type:        "object",
			Description: "Security configuration settings",
			Required:    false,
		},
		{
			Name:        "performance",
			Type:        "object",
			Description: "Performance tuning settings",
			Required:    false,
		},
		{
			Name:        "options",
			Type:        "object",
			Description: "Additional options and configuration",
			Required:    false,
		},
	}

	// Create base function
	baseFunc := functionbase.NewBaseFunction(
		"machines",
		"Manage machine configurations, hardware detection, and system optimization for NixOS machines",
		parameters,
	)

	// Add examples
	baseFunc.SetSchema(functionbase.FunctionSchema{
		Name:        "machines",
		Description: "Manage machine configurations, hardware detection, and system optimization for NixOS machines",
		Parameters:  parameters,
		Examples: []functionbase.FunctionExample{
			{
				Description: "List available machine configurations",
				Parameters: map[string]interface{}{
					"operation": "list",
				},
				Expected: "Returns a list of available machine configurations and templates",
			},
			{
				Description: "Create a new desktop machine configuration",
				Parameters: map[string]interface{}{
					"operation":    "create",
					"machine_name": "desktop-workstation",
					"machine_type": "desktop",
					"architecture": "x86_64",
					"environment":  "development",
					"services":     []string{"xserver", "pipewire", "networkmanager"},
				},
				Expected: "Returns a complete machine configuration for a desktop workstation",
			},
			{
				Description: "Analyze hardware configuration",
				Parameters: map[string]interface{}{
					"operation":    "hardware",
					"machine_name": "current-system",
				},
				Expected: "Returns detailed hardware information and optimization recommendations",
			},
		},
	})

	return &MachinesFunction{
		BaseFunction:  baseFunc,
		machinesAgent: nil, // Set to nil to avoid provider requirement
		logger:        logger.NewLogger(),
	}
}

// Execute performs the machines operation
func (f *MachinesFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	start := time.Now()
	f.logger.Info("Executing machines function")

	// Parse parameters
	var req MachinesRequest
	if err := f.parseParameters(params, &req); err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("parameter parsing failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	// Validate parameters
	if err := f.ValidateParameters(params); err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("parameter validation failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	// Execute operation based on type
	var response *MachinesResponse
	var err error

	switch req.Operation {
	case "list":
		response, err = f.executeListing(ctx, &req)
	case "create":
		response, err = f.executeCreation(ctx, &req)
	case "configure":
		response, err = f.executeConfiguration(ctx, &req)
	case "template":
		response, err = f.executeTemplate(ctx, &req)
	case "analyze":
		response, err = f.executeAnalysis(ctx, &req)
	case "optimize":
		response, err = f.executeOptimization(ctx, &req)
	case "security":
		response, err = f.executeSecurity(ctx, &req)
	case "performance":
		response, err = f.executePerformance(ctx, &req)
	case "network":
		response, err = f.executeNetwork(ctx, &req)
	case "hardware":
		response, err = f.executeHardware(ctx, &req)
	default:
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("unsupported operation: %s", req.Operation),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	if err != nil {
		return &functionbase.FunctionResult{
			Success:   false,
			Error:     fmt.Sprintf("operation failed: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, err
	}

	return &functionbase.FunctionResult{
		Success:   true,
		Data:      response,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// parseParameters parses the input parameters
func (f *MachinesFunction) parseParameters(params map[string]interface{}, req *MachinesRequest) error {
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, req); err != nil {
		return fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return nil
}

// formatServices formats a list of services for NixOS configuration
func (f *MachinesFunction) formatServices(services []string) string {
	if len(services) == 0 {
		return ""
	}

	var formatted []string
	for _, service := range services {
		switch service {
		case "xserver":
			formatted = append(formatted, "    xserver.enable = true;")
		case "pipewire":
			formatted = append(formatted, "    pipewire.enable = true;")
		case "networkmanager":
			formatted = append(formatted, "    networkmanager.enable = true;")
		case "openssh":
			formatted = append(formatted, "    openssh.enable = true;")
		case "nginx":
			formatted = append(formatted, "    nginx.enable = true;")
		case "postgresql":
			formatted = append(formatted, "    postgresql.enable = true;")
		default:
			formatted = append(formatted, fmt.Sprintf("    %s.enable = true;", service))
		}
	}

	return strings.Join(formatted, "\n")
}

// executeListing handles machine listing operations
func (f *MachinesFunction) executeListing(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines listing operation")

	// Create a list of available machine configurations and templates
	machines := []MachineInfo{
		{
			Name:         "desktop-workstation",
			Type:         "desktop",
			Architecture: "x86_64",
			Status:       "available",
			Environment:  "development",
			Services:     []string{"xserver", "pipewire", "networkmanager"},
		},
		{
			Name:         "server-minimal",
			Type:         "server",
			Architecture: "x86_64",
			Status:       "available",
			Environment:  "production",
			Services:     []string{"openssh", "nginx", "postgresql"},
		},
	}

	templates := []MachineTemplate{
		{
			Name:         "desktop-dev",
			Description:  "Desktop workstation for development",
			Type:         "desktop",
			Architecture: "x86_64",
			Category:     "development",
			Tags:         []string{"development", "gui", "programming"},
			Features:     []string{"IDE support", "multiple monitors", "development tools"},
		},
		{
			Name:         "server-web",
			Description:  "Web server configuration",
			Type:         "server",
			Architecture: "x86_64",
			Category:     "server",
			Tags:         []string{"web", "server", "production"},
			Features:     []string{"nginx", "ssl", "database"},
		},
	}

	return &MachinesResponse{
		Operation: "list",
		Status:    "success",
		Machines:  machines,
		Templates: templates,
	}, nil
}

// executeCreation handles machine creation operations
func (f *MachinesFunction) executeCreation(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines creation operation")

	// Create a basic machine configuration
	config := &MachineConfiguration{
		Name:         req.MachineName,
		Type:         req.MachineType,
		Architecture: req.Architecture,
		Services:     req.Services,
		Hardware:     make(map[string]interface{}),
		Network:      make(map[string]interface{}),
		Security:     make(map[string]interface{}),
		Performance:  make(map[string]interface{}),
		Environment:  make(map[string]string),
		Files:        make(map[string]string),
	}

	// Add basic configuration
	if req.Configuration != nil {
		for k, v := range req.Configuration {
			config.Environment[k] = v
		}
	}

	setupSteps := []string{
		"1. Create configuration.nix with the specified settings",
		"2. Configure hardware detection and drivers",
		"3. Set up networking and firewall",
		"4. Install and configure specified services",
		"5. Apply security hardening",
		"6. Optimize performance settings",
		"7. Test the configuration",
		"8. Rebuild and switch to new configuration",
	}

	return &MachinesResponse{
		Operation:     "create",
		Status:        "success",
		Configuration: config,
		SetupSteps:    setupSteps,
	}, nil
}

// executeConfiguration handles machine configuration operations
func (f *MachinesFunction) executeConfiguration(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines configuration operation")

	config := &MachineConfiguration{
		Name:         req.MachineName,
		Type:         req.MachineType,
		Architecture: req.Architecture,
		Services:     req.Services,
		Hardware:     make(map[string]interface{}),
		Network:      make(map[string]interface{}),
		Security:     make(map[string]interface{}),
		Performance:  make(map[string]interface{}),
		Environment:  make(map[string]string),
		Files:        make(map[string]string),
	}

	// Generate NixOS configuration
	nixConfig := fmt.Sprintf(`{ config, pkgs, ... }:

{
  # Machine: %s
  # Type: %s
  # Architecture: %s

  imports = [
    ./hardware-configuration.nix
  ];

  # System settings
  system.stateVersion = "24.05";
  
  # Services
  services = {
%s
  };

  # Environment
  environment.systemPackages = with pkgs; [
    vim
    git
    htop
  ];
}`, req.MachineName, req.MachineType, req.Architecture, f.formatServices(req.Services))

	config.Configuration = nixConfig

	return &MachinesResponse{
		Operation:     "configure",
		Status:        "success",
		Configuration: config,
	}, nil
}

// executeTemplate handles machine template operations
func (f *MachinesFunction) executeTemplate(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines template operation")

	templates := []MachineTemplate{
		{
			Name:         "desktop-gaming",
			Description:  "Gaming desktop with GPU support",
			Type:         "desktop",
			Architecture: "x86_64",
			Category:     "gaming",
			Tags:         []string{"gaming", "gpu", "performance"},
			Features:     []string{"Steam", "GPU drivers", "high performance"},
			UseCases:     []string{"Gaming", "Content creation", "Streaming"},
		},
		{
			Name:         "server-homelab",
			Description:  "Home lab server configuration",
			Type:         "server",
			Architecture: "x86_64",
			Category:     "homelab",
			Tags:         []string{"homelab", "self-hosted", "containers"},
			Features:     []string{"Docker", "Reverse proxy", "Monitoring"},
			UseCases:     []string{"Self-hosting", "Learning", "Development"},
		},
	}

	return &MachinesResponse{
		Operation: "template",
		Status:    "success",
		Templates: templates,
	}, nil
}

// executeAnalysis handles machine analysis operations
func (f *MachinesFunction) executeAnalysis(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines analysis operation")

	requirements := []string{
		"NixOS 23.11 or later",
		"Sufficient disk space (minimum 20GB)",
		"Network connectivity for package downloads",
		"Hardware compatibility check completed",
	}

	recommendations := []string{
		"Enable automatic garbage collection",
		"Configure binary cache for faster builds",
		"Set up regular system backups",
		"Monitor system performance",
		"Keep system updated",
	}

	hardwareInfo := &HardwareInfo{
		CPU: &CPUInfo{
			Model:        "Detected automatically",
			Cores:        0,
			Threads:      0,
			Architecture: req.Architecture,
			Features:     []string{"Hardware detection required"},
		},
		Memory: &MemoryInfo{
			Total:     "To be detected",
			Available: "To be detected",
			Type:      "DDR4/DDR5",
		},
	}

	return &MachinesResponse{
		Operation:       "analyze",
		Status:          "success",
		Requirements:    requirements,
		Recommendations: recommendations,
		HardwareInfo:    hardwareInfo,
	}, nil
}

// executeOptimization handles machine optimization operations
func (f *MachinesFunction) executeOptimization(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines optimization operation")

	if f.machinesAgent == nil {
		// Return static optimization recommendations
		performanceConfig := []PerformanceSetting{
			{
				Name:        "kernel.performance",
				Value:       "true",
				Description: "Enable performance-oriented kernel settings",
				Category:    "kernel",
				Impact:      "high",
				Recommended: true,
			},
			{
				Name:        "boot.loader.timeout",
				Value:       "1",
				Description: "Reduce boot loader timeout for faster startup",
				Category:    "boot",
				Impact:      "low",
				Recommended: true,
			},
		}

		recommendations := []string{
			"Enable automatic garbage collection",
			"Configure binary cache for faster builds",
			"Optimize kernel parameters for performance",
			"Enable SSD optimizations if applicable",
			"Configure swap appropriately",
		}

		return &MachinesResponse{
			Operation:         "optimize",
			Status:            "success",
			PerformanceConfig: performanceConfig,
			Recommendations:   recommendations,
		}, nil
	}

	// Build prompt for optimization
	prompt := fmt.Sprintf(`Optimize the NixOS machine configuration for performance and efficiency.

Machine: %s (Type: %s, Architecture: %s)
Environment: %s
Current Configuration: %v
Hardware Info: %v
Performance Settings: %v

Please provide:
1. Performance optimization recommendations
2. Resource usage optimization
3. System tuning suggestions
4. Hardware-specific optimizations
5. Configuration improvements
6. Best practices for the target environment

Focus on practical, implementable optimizations.`,
		req.MachineName, req.MachineType, req.Architecture, req.Environment,
		req.Configuration, req.Hardware, req.Performance)

	// Query the agent
	agentResponse, err := f.machinesAgent.GenerateResponse(ctx, prompt)
	if err != nil {
		return &MachinesResponse{
			Operation: "optimize",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to optimize machine: %v", err),
		}, nil
	}

	// Parse the response
	response, err := f.parseAgentResponse(agentResponse, "optimize")
	if err != nil {
		return &MachinesResponse{
			Operation: "optimize",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to parse agent response: %v", err),
		}, nil
	}

	response.Operation = req.Operation
	response.Status = "success"
	return response, nil
}

// executeSecurity handles machine security operations
func (f *MachinesFunction) executeSecurity(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines security operation")

	if f.machinesAgent == nil {
		return &MachinesResponse{
			Operation: "security",
			Status:    "error",
			Error:     "No machine agent available. This function requires a provider to be configured.",
		}, nil
	}

	// Build prompt for security configuration
	prompt := fmt.Sprintf(`Configure security settings for NixOS machine.

Machine: %s (Type: %s, Architecture: %s)
Environment: %s
Current Configuration: %v
Security Requirements: %v

Please provide:
1. Security hardening recommendations
2. Firewall configuration
3. Access control settings
4. Encryption and privacy settings
5. Security monitoring recommendations
6. Best practices for the target environment

Focus on practical, implementable security measures.`,
		req.MachineName, req.MachineType, req.Architecture, req.Environment,
		req.Configuration, req.Security)

	// Query the agent
	agentResponse, err := f.machinesAgent.GenerateResponse(ctx, prompt)
	if err != nil {
		return &MachinesResponse{
			Operation: "security",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to configure security: %v", err),
		}, nil
	}

	// Parse the response
	response, err := f.parseAgentResponse(agentResponse, "security")
	if err != nil {
		return &MachinesResponse{
			Operation: "security",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to parse agent response: %v", err),
		}, nil
	}

	response.Operation = req.Operation
	response.Status = "success"
	return response, nil
}

// executePerformance handles machine performance operations
func (f *MachinesFunction) executePerformance(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines performance operation")

	if f.machinesAgent == nil {
		return &MachinesResponse{
			Operation: "performance",
			Status:    "error",
			Error:     "No machine agent available. This function requires a provider to be configured.",
		}, nil
	}

	// Build prompt for performance optimization
	prompt := fmt.Sprintf(`Optimize performance for NixOS machine.

Machine: %s (Type: %s, Architecture: %s)
Environment: %s
Current Configuration: %v
Hardware Info: %v
Performance Settings: %v

Please provide:
1. Performance optimization recommendations
2. Resource allocation tuning
3. System performance settings
4. Hardware-specific optimizations
5. Service performance tuning
6. Best practices for the target environment

Focus on practical, implementable performance improvements.`,
		req.MachineName, req.MachineType, req.Architecture, req.Environment,
		req.Configuration, req.Hardware, req.Performance)

	// Query the agent
	agentResponse, err := f.machinesAgent.GenerateResponse(ctx, prompt)
	if err != nil {
		return &MachinesResponse{
			Operation: "performance",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to optimize performance: %v", err),
		}, nil
	}

	// Parse the response
	response, err := f.parseAgentResponse(agentResponse, "performance")
	if err != nil {
		return &MachinesResponse{
			Operation: "performance",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to parse agent response: %v", err),
		}, nil
	}

	response.Operation = req.Operation
	response.Status = "success"
	return response, nil
}

// executeNetwork handles machine network operations
func (f *MachinesFunction) executeNetwork(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines network operation")

	if f.machinesAgent == nil {
		return &MachinesResponse{
			Operation: "network",
			Status:    "error",
			Error:     "No machine agent available. This function requires a provider to be configured.",
		}, nil
	}

	// Build prompt for network configuration
	prompt := fmt.Sprintf(`Configure network settings for NixOS machine.

Machine: %s (Type: %s, Architecture: %s)
Environment: %s
Current Configuration: %v
Network Requirements: %v
Security Settings: %v

Please provide:
1. Network interface configuration
2. Routing and DNS settings
3. Firewall configuration
4. VPN setup if needed
5. Network optimization recommendations
6. Security considerations for network setup

Focus on practical, implementable network configurations.`,
		req.MachineName, req.MachineType, req.Architecture, req.Environment,
		req.Configuration, req.Network, req.Security)

	// Query the agent
	agentResponse, err := f.machinesAgent.GenerateResponse(ctx, prompt)
	if err != nil {
		return &MachinesResponse{
			Operation: "network",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to configure network: %v", err),
		}, nil
	}

	// Parse the response
	response, err := f.parseAgentResponse(agentResponse, "network")
	if err != nil {
		return &MachinesResponse{
			Operation: "network",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to parse agent response: %v", err),
		}, nil
	}

	response.Operation = req.Operation
	response.Status = "success"
	return response, nil
}

// executeHardware handles machine hardware operations
func (f *MachinesFunction) executeHardware(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines hardware operation")

	if f.machinesAgent == nil {
		return &MachinesResponse{
			Operation: "hardware",
			Status:    "error",
			Error:     "No machine agent available. This function requires a provider to be configured.",
		}, nil
	}

	// Build prompt for hardware analysis
	prompt := fmt.Sprintf(`Analyze hardware configuration for NixOS machine.

Machine: %s (Type: %s, Architecture: %s)
Environment: %s
Current Configuration: %v
Hardware Info: %v

Please provide:
1. Hardware compatibility analysis
2. Driver recommendations
3. Hardware optimization settings
4. Performance tuning for detected hardware
5. Potential hardware issues and solutions
6. NixOS-specific hardware configuration

Focus on practical hardware configuration recommendations.`,
		req.MachineName, req.MachineType, req.Architecture, req.Environment,
		req.Configuration, req.Hardware)

	// Query the agent
	agentResponse, err := f.machinesAgent.GenerateResponse(ctx, prompt)
	if err != nil {
		return &MachinesResponse{
			Operation: "hardware",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to analyze hardware: %v", err),
		}, nil
	}

	// Parse the response
	response, err := f.parseAgentResponse(agentResponse, "hardware")
	if err != nil {
		return &MachinesResponse{
			Operation: "hardware",
			Status:    "error",
			Error:     fmt.Sprintf("Failed to parse agent response: %v", err),
		}, nil
	}

	response.Operation = req.Operation
	response.Status = "success"
	return response, nil
}

// parseAgentResponse parses the agent response into a MachinesResponse
func (f *MachinesFunction) parseAgentResponse(agentResponse string, operation string) (*MachinesResponse, error) {
	response := &MachinesResponse{
		Operation: operation,
		Status:    "processing",
	}

	// Try to parse as JSON first
	var jsonResponse MachinesResponse
	if err := json.Unmarshal([]byte(agentResponse), &jsonResponse); err == nil {
		return &jsonResponse, nil
	}

	// Parse text response based on operation type
	switch operation {
	case "list":
		response.Machines = f.extractMachineList(agentResponse)
		response.Templates = f.extractTemplates(agentResponse)
	case "create", "configure":
		response.Configuration = f.extractConfiguration(agentResponse)
		response.SetupSteps = f.extractSetupSteps(agentResponse)
		response.Commands = f.extractCommands(agentResponse)
	case "template":
		response.Templates = f.extractTemplates(agentResponse)
		response.Examples = f.extractExamples(agentResponse)
	case "analyze":
		response.Requirements = f.extractRequirements(agentResponse)
		response.Recommendations = f.extractRecommendations(agentResponse)
		response.HardwareInfo = f.extractHardwareInfo(agentResponse)
	case "optimize":
		response.PerformanceConfig = f.extractPerformanceConfig(agentResponse)
		response.Recommendations = f.extractRecommendations(agentResponse)
	case "security":
		response.SecuritySettings = f.extractSecuritySettings(agentResponse)
		response.Recommendations = f.extractRecommendations(agentResponse)
	case "performance":
		response.PerformanceConfig = f.extractPerformanceConfig(agentResponse)
		response.Recommendations = f.extractRecommendations(agentResponse)
	case "network":
		response.NetworkConfig = f.extractNetworkConfig(agentResponse)
		response.SetupSteps = f.extractSetupSteps(agentResponse)
	case "hardware":
		response.HardwareInfo = f.extractHardwareInfo(agentResponse)
		response.Recommendations = f.extractRecommendations(agentResponse)
	}

	// Extract common elements
	response.Documentation = f.extractDocumentation(agentResponse)

	return response, nil
}

// Helper functions for parsing agent responses
func (f *MachinesFunction) extractMachineList(response string) []MachineInfo {
	// Implementation would parse machine information from response
	return []MachineInfo{}
}

func (f *MachinesFunction) extractTemplates(response string) []MachineTemplate {
	// Implementation would parse template information from response
	return []MachineTemplate{}
}

func (f *MachinesFunction) extractConfiguration(response string) *MachineConfiguration {
	// Implementation would parse configuration from response
	return &MachineConfiguration{}
}

func (f *MachinesFunction) extractSetupSteps(response string) []string {
	// Implementation would parse setup steps from response
	return []string{}
}

func (f *MachinesFunction) extractCommands(response string) []string {
	// Implementation would parse commands from response
	return []string{}
}

func (f *MachinesFunction) extractRequirements(response string) []string {
	// Implementation would parse requirements from response
	return []string{}
}

func (f *MachinesFunction) extractRecommendations(response string) []string {
	// Implementation would parse recommendations from response
	return []string{}
}

func (f *MachinesFunction) extractHardwareInfo(response string) *HardwareInfo {
	// Implementation would parse hardware information from response
	return &HardwareInfo{}
}

func (f *MachinesFunction) extractPerformanceConfig(response string) []PerformanceSetting {
	// Implementation would parse performance configuration from response
	return []PerformanceSetting{}
}

func (f *MachinesFunction) extractSecuritySettings(response string) []SecuritySetting {
	// Implementation would parse security settings from response
	return []SecuritySetting{}
}

func (f *MachinesFunction) extractNetworkConfig(response string) *NetworkConfiguration {
	// Implementation would parse network configuration from response
	return &NetworkConfiguration{}
}

func (f *MachinesFunction) extractDocumentation(response string) []DocumentationLink {
	// Implementation would parse documentation links from response
	return []DocumentationLink{}
}

func (f *MachinesFunction) extractExamples(response string) []ConfigurationExample {
	// Implementation would parse examples from response
	return []ConfigurationExample{}
}
