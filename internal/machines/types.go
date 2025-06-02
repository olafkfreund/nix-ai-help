package machines

import (
	"time"
)

// Machine represents a NixOS machine in the registry
type Machine struct {
	Name        string            `yaml:"name" json:"name"`
	Host        string            `yaml:"host" json:"host"`
	Port        int               `yaml:"port" json:"port"`
	User        string            `yaml:"user" json:"user"`
	SSHKey      string            `yaml:"ssh_key,omitempty" json:"ssh_key,omitempty"`
	NixOSPath   string            `yaml:"nixos_path" json:"nixos_path"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Groups      []string          `yaml:"groups,omitempty" json:"groups,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	LastSync    *time.Time        `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`
	LastDeploy  *time.Time        `yaml:"last_deploy,omitempty" json:"last_deploy,omitempty"`
	Status      MachineStatus     `yaml:"status" json:"status"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
}

// MachineStatus represents the current status of a machine
type MachineStatus string

const (
	StatusOnline    MachineStatus = "online"
	StatusOffline   MachineStatus = "offline"
	StatusSyncing   MachineStatus = "syncing"
	StatusDeploying MachineStatus = "deploying"
	StatusError     MachineStatus = "error"
	StatusUnknown   MachineStatus = "unknown"
)

// MachineGroup represents a group of machines for fleet management
type MachineGroup struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	Machines    []string  `yaml:"machines" json:"machines"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}

// Registry holds all machines and groups
type Registry struct {
	Machines  []Machine      `yaml:"machines" json:"machines"`
	Groups    []MachineGroup `yaml:"groups" json:"groups"`
	Version   string         `yaml:"version" json:"version"`
	UpdatedAt time.Time      `yaml:"updated_at" json:"updated_at"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Machine   string        `json:"machine"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	FilesSync int           `json:"files_synced"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// DeployResult represents the result of a deployment operation
type DeployResult struct {
	Machine     string        `json:"machine"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Generation  int           `json:"generation,omitempty"`
	Duration    time.Duration `json:"duration"`
	Timestamp   time.Time     `json:"timestamp"`
	RollbackCmd string        `json:"rollback_cmd,omitempty"`
}

// DiffResult represents configuration differences between machines
type DiffResult struct {
	MachineA    string    `json:"machine_a"`
	MachineB    string    `json:"machine_b"`
	Differences []string  `json:"differences"`
	Summary     string    `json:"summary"`
	Timestamp   time.Time `json:"timestamp"`
}

// ConnectionInfo contains SSH connection details
type ConnectionInfo struct {
	Host    string
	Port    int
	User    string
	SSHKey  string
	Timeout time.Duration
}

// GetConnectionInfo returns SSH connection information for the machine
func (m *Machine) GetConnectionInfo() ConnectionInfo {
	port := m.Port
	if port == 0 {
		port = 22 // Default SSH port
	}

	user := m.User
	if user == "" {
		user = "root" // Default NixOS user for system operations
	}

	return ConnectionInfo{
		Host:    m.Host,
		Port:    port,
		User:    user,
		SSHKey:  m.SSHKey,
		Timeout: 30 * time.Second,
	}
}

// IsOnline checks if the machine is currently online
func (m *Machine) IsOnline() bool {
	return m.Status == StatusOnline
}

// CanDeploy checks if the machine is ready for deployment
func (m *Machine) CanDeploy() bool {
	return m.Status == StatusOnline || m.Status == StatusUnknown
}

// AddTag adds a tag to the machine if it doesn't already exist
func (m *Machine) AddTag(tag string) {
	for _, t := range m.Tags {
		if t == tag {
			return
		}
	}
	m.Tags = append(m.Tags, tag)
}

// RemoveTag removes a tag from the machine
func (m *Machine) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)
			return
		}
	}
}

// HasTag checks if the machine has a specific tag
func (m *Machine) HasTag(tag string) bool {
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
