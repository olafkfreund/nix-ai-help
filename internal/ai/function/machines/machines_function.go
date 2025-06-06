package machines

import (
	"context"
	"encoding/json"
	"fmt"
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
		machinesAgent: agent.NewMachinesAgent(),
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

// executeListing handles machine listing operations
func (f *MachinesFunction) executeListing(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines listing operation")

	// Use agent to get machine information
	machineContext := &agent.MachinesContext{
		Operation:    req.Operation,
		MachineName:  req.MachineName,
		MachineType:  req.MachineType,
		Architecture: req.Architecture,
		Environment:  req.Environment,
	}

	agentResponse, err := f.machinesAgent.ListMachines(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent listing failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeCreation handles machine creation operations
func (f *MachinesFunction) executeCreation(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines creation operation")

	// Use agent to create machine configuration
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Services:      req.Services,
		Configuration: req.Configuration,
		Template:      req.Template,
		Hardware:      req.Hardware,
		Network:       req.Network,
		Security:      req.Security,
		Performance:   req.Performance,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.CreateMachine(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent creation failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeConfiguration handles machine configuration operations
func (f *MachinesFunction) executeConfiguration(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines configuration operation")

	// Use agent to configure machine
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Services:      req.Services,
		Configuration: req.Configuration,
		Hardware:      req.Hardware,
		Network:       req.Network,
		Security:      req.Security,
		Performance:   req.Performance,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.ConfigureMachine(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent configuration failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeTemplate handles machine template operations
func (f *MachinesFunction) executeTemplate(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines template operation")

	// Use agent to get templates
	machineContext := &agent.MachinesContext{
		Operation:    req.Operation,
		MachineName:  req.MachineName,
		MachineType:  req.MachineType,
		Architecture: req.Architecture,
		Environment:  req.Environment,
		Template:     req.Template,
		Options:      req.Options,
	}

	agentResponse, err := f.machinesAgent.GetTemplates(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent template operation failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeAnalysis handles machine analysis operations
func (f *MachinesFunction) executeAnalysis(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines analysis operation")

	// Use agent to analyze machine
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Hardware:      req.Hardware,
		Network:       req.Network,
		Security:      req.Security,
		Performance:   req.Performance,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.AnalyzeMachine(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent analysis failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeOptimization handles machine optimization operations
func (f *MachinesFunction) executeOptimization(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines optimization operation")

	// Use agent to optimize machine
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Hardware:      req.Hardware,
		Network:       req.Network,
		Security:      req.Security,
		Performance:   req.Performance,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.OptimizeMachine(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent optimization failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeSecurity handles machine security operations
func (f *MachinesFunction) executeSecurity(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines security operation")

	// Use agent to configure security
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Security:      req.Security,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.ConfigureSecurity(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent security configuration failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executePerformance handles machine performance operations
func (f *MachinesFunction) executePerformance(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines performance operation")

	// Use agent to optimize performance
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Hardware:      req.Hardware,
		Performance:   req.Performance,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.OptimizePerformance(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent performance optimization failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeNetwork handles machine network operations
func (f *MachinesFunction) executeNetwork(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines network operation")

	// Use agent to configure network
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Network:       req.Network,
		Security:      req.Security,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.ConfigureNetwork(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent network configuration failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	response.Operation = req.Operation
	response.Status = "success"

	return response, nil
}

// executeHardware handles machine hardware operations
func (f *MachinesFunction) executeHardware(ctx context.Context, req *MachinesRequest) (*MachinesResponse, error) {
	f.logger.Info("Executing machines hardware operation")

	// Use agent to analyze hardware
	machineContext := &agent.MachinesContext{
		Operation:     req.Operation,
		MachineName:   req.MachineName,
		MachineType:   req.MachineType,
		Architecture:  req.Architecture,
		Environment:   req.Environment,
		Configuration: req.Configuration,
		Hardware:      req.Hardware,
		Options:       req.Options,
	}

	agentResponse, err := f.machinesAgent.AnalyzeHardware(machineContext)
	if err != nil {
		return nil, fmt.Errorf("agent hardware analysis failed: %w", err)
	}

	// Parse the agent response
	response, err := f.parseAgentResponse(agentResponse, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
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
