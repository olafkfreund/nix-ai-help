package machines

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"gopkg.in/yaml.v3"
)

const (
	registryFileName = "machines.yaml"
	registryVersion  = "1.0"
)

var log = logger.NewLogger()

// RegistryManager handles machine registry operations
type RegistryManager struct {
	registryPath string
	registry     *Registry
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() (*RegistryManager, error) {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	registryPath := filepath.Join(configDir, registryFileName)

	rm := &RegistryManager{
		registryPath: registryPath,
	}

	// Load existing registry or create new one
	if err := rm.loadRegistry(); err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	return rm, nil
}

// loadRegistry loads the registry from disk or creates a new one
func (rm *RegistryManager) loadRegistry() error {
	if !utils.IsFile(rm.registryPath) {
		// Create new registry
		rm.registry = &Registry{
			Machines:  make([]Machine, 0),
			Groups:    make([]MachineGroup, 0),
			Version:   registryVersion,
			UpdatedAt: time.Now(),
		}
		return rm.saveRegistry()
	}

	data, err := os.ReadFile(rm.registryPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	registry := &Registry{}
	if err := yaml.Unmarshal(data, registry); err != nil {
		return fmt.Errorf("failed to parse registry file: %w", err)
	}

	rm.registry = registry
	return nil
}

// saveRegistry saves the registry to disk
func (rm *RegistryManager) saveRegistry() error {
	rm.registry.UpdatedAt = time.Now()

	data, err := yaml.Marshal(rm.registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(rm.registryPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(rm.registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	log.Debug(fmt.Sprintf("Registry saved to %s", rm.registryPath))
	return nil
}

// ListMachines returns all registered machines
func (rm *RegistryManager) ListMachines() []Machine {
	return rm.registry.Machines
}

// ListGroups returns all machine groups
func (rm *RegistryManager) ListGroups() []MachineGroup {
	return rm.registry.Groups
}

// GetMachine returns a machine by name
func (rm *RegistryManager) GetMachine(name string) (*Machine, error) {
	for i := range rm.registry.Machines {
		if rm.registry.Machines[i].Name == name {
			return &rm.registry.Machines[i], nil
		}
	}
	return nil, fmt.Errorf("machine '%s' not found", name)
}

// GetGroup returns a group by name
func (rm *RegistryManager) GetGroup(name string) (*MachineGroup, error) {
	for i := range rm.registry.Groups {
		if rm.registry.Groups[i].Name == name {
			return &rm.registry.Groups[i], nil
		}
	}
	return nil, fmt.Errorf("group '%s' not found", name)
}

// AddMachine adds a new machine to the registry
func (rm *RegistryManager) AddMachine(machine Machine) error {
	// Check if machine already exists
	for _, m := range rm.registry.Machines {
		if m.Name == machine.Name {
			return fmt.Errorf("machine '%s' already exists", machine.Name)
		}
		if m.Host == machine.Host {
			return fmt.Errorf("machine with host '%s' already exists as '%s'", machine.Host, m.Name)
		}
	}

	// Set timestamps
	now := time.Now()
	machine.CreatedAt = now
	machine.UpdatedAt = now
	machine.Status = StatusUnknown

	// Add to registry
	rm.registry.Machines = append(rm.registry.Machines, machine)

	log.Info(fmt.Sprintf("Added machine '%s' (%s) to registry", machine.Name, machine.Host))
	return rm.saveRegistry()
}

// UpdateMachine updates an existing machine
func (rm *RegistryManager) UpdateMachine(name string, updates Machine) error {
	for i := range rm.registry.Machines {
		if rm.registry.Machines[i].Name == name {
			// Preserve creation time and update timestamp
			updates.CreatedAt = rm.registry.Machines[i].CreatedAt
			updates.UpdatedAt = time.Now()
			updates.Name = name // Ensure name doesn't change

			rm.registry.Machines[i] = updates
			log.Info(fmt.Sprintf("Updated machine '%s'", name))
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("machine '%s' not found", name)
}

// RemoveMachine removes a machine from the registry
func (rm *RegistryManager) RemoveMachine(name string) error {
	for i, machine := range rm.registry.Machines {
		if machine.Name == name {
			// Remove from all groups
			for j := range rm.registry.Groups {
				rm.removeMachineFromGroup(&rm.registry.Groups[j], name)
			}

			// Remove from machines list
			rm.registry.Machines = append(rm.registry.Machines[:i], rm.registry.Machines[i+1:]...)

			log.Info(fmt.Sprintf("Removed machine '%s' from registry", name))
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("machine '%s' not found", name)
}

// AddGroup adds a new machine group
func (rm *RegistryManager) AddGroup(group MachineGroup) error {
	// Check if group already exists
	for _, g := range rm.registry.Groups {
		if g.Name == group.Name {
			return fmt.Errorf("group '%s' already exists", group.Name)
		}
	}

	// Validate that all machines exist
	for _, machineName := range group.Machines {
		if _, err := rm.GetMachine(machineName); err != nil {
			return fmt.Errorf("machine '%s' in group does not exist", machineName)
		}
	}

	// Set timestamps
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	rm.registry.Groups = append(rm.registry.Groups, group)

	log.Info(fmt.Sprintf("Added group '%s' with %d machines", group.Name, len(group.Machines)))
	return rm.saveRegistry()
}

// UpdateGroup updates an existing group
func (rm *RegistryManager) UpdateGroup(name string, updates MachineGroup) error {
	for i := range rm.registry.Groups {
		if rm.registry.Groups[i].Name == name {
			// Validate that all machines exist
			for _, machineName := range updates.Machines {
				if _, err := rm.GetMachine(machineName); err != nil {
					return fmt.Errorf("machine '%s' in group does not exist", machineName)
				}
			}

			// Preserve creation time and update timestamp
			updates.CreatedAt = rm.registry.Groups[i].CreatedAt
			updates.UpdatedAt = time.Now()
			updates.Name = name // Ensure name doesn't change

			rm.registry.Groups[i] = updates
			log.Info(fmt.Sprintf("Updated group '%s'", name))
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("group '%s' not found", name)
}

// RemoveGroup removes a group from the registry
func (rm *RegistryManager) RemoveGroup(name string) error {
	for i, group := range rm.registry.Groups {
		if group.Name == name {
			rm.registry.Groups = append(rm.registry.Groups[:i], rm.registry.Groups[i+1:]...)
			log.Info(fmt.Sprintf("Removed group '%s' from registry", name))
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("group '%s' not found", name)
}

// AddMachineToGroup adds a machine to a group
func (rm *RegistryManager) AddMachineToGroup(groupName, machineName string) error {
	// Verify machine exists
	if _, err := rm.GetMachine(machineName); err != nil {
		return err
	}

	group, err := rm.GetGroup(groupName)
	if err != nil {
		return err
	}

	// Check if machine is already in group
	for _, m := range group.Machines {
		if m == machineName {
			return fmt.Errorf("machine '%s' is already in group '%s'", machineName, groupName)
		}
	}

	group.Machines = append(group.Machines, machineName)
	return rm.UpdateGroup(groupName, *group)
}

// RemoveMachineFromGroup removes a machine from a group
func (rm *RegistryManager) RemoveMachineFromGroup(groupName, machineName string) error {
	group, err := rm.GetGroup(groupName)
	if err != nil {
		return err
	}

	if rm.removeMachineFromGroup(group, machineName) {
		return rm.UpdateGroup(groupName, *group)
	}

	return fmt.Errorf("machine '%s' not found in group '%s'", machineName, groupName)
}

// removeMachineFromGroup is a helper function to remove a machine from a group
func (rm *RegistryManager) removeMachineFromGroup(group *MachineGroup, machineName string) bool {
	for i, m := range group.Machines {
		if m == machineName {
			group.Machines = append(group.Machines[:i], group.Machines[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateMachineStatus updates the status of a machine
func (rm *RegistryManager) UpdateMachineStatus(name string, status MachineStatus) error {
	for i := range rm.registry.Machines {
		if rm.registry.Machines[i].Name == name {
			rm.registry.Machines[i].Status = status
			rm.registry.Machines[i].UpdatedAt = time.Now()
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("machine '%s' not found", name)
}

// UpdateLastSync updates the last sync time for a machine
func (rm *RegistryManager) UpdateLastSync(name string) error {
	for i := range rm.registry.Machines {
		if rm.registry.Machines[i].Name == name {
			now := time.Now()
			rm.registry.Machines[i].LastSync = &now
			rm.registry.Machines[i].UpdatedAt = now
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("machine '%s' not found", name)
}

// UpdateLastDeploy updates the last deploy time for a machine
func (rm *RegistryManager) UpdateLastDeploy(name string) error {
	for i := range rm.registry.Machines {
		if rm.registry.Machines[i].Name == name {
			now := time.Now()
			rm.registry.Machines[i].LastDeploy = &now
			rm.registry.Machines[i].UpdatedAt = now
			return rm.saveRegistry()
		}
	}
	return fmt.Errorf("machine '%s' not found", name)
}

// GetMachinesByTag returns all machines that have a specific tag
func (rm *RegistryManager) GetMachinesByTag(tag string) []Machine {
	var machines []Machine
	for _, machine := range rm.registry.Machines {
		if machine.HasTag(tag) {
			machines = append(machines, machine)
		}
	}
	return machines
}

// GetMachinesByGroup returns all machines in a specific group
func (rm *RegistryManager) GetMachinesByGroup(groupName string) ([]Machine, error) {
	group, err := rm.GetGroup(groupName)
	if err != nil {
		return nil, err
	}

	var machines []Machine
	for _, machineName := range group.Machines {
		if machine, err := rm.GetMachine(machineName); err == nil {
			machines = append(machines, *machine)
		}
	}

	return machines, nil
}

// GetRegistryPath returns the path to the registry file
func (rm *RegistryManager) GetRegistryPath() string {
	return rm.registryPath
}

// ExportRegistry exports the registry to a file
func (rm *RegistryManager) ExportRegistry(path string) error {
	data, err := yaml.Marshal(rm.registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	log.Info(fmt.Sprintf("Registry exported to %s", path))
	return nil
}

// ImportRegistry imports a registry from a file
func (rm *RegistryManager) ImportRegistry(path string, merge bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	importedRegistry := &Registry{}
	if err := yaml.Unmarshal(data, importedRegistry); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	if merge {
		// Merge imported registry with existing
		return rm.mergeRegistry(importedRegistry)
	} else {
		// Replace existing registry
		rm.registry = importedRegistry
		return rm.saveRegistry()
	}
}

// mergeRegistry merges an imported registry with the existing one
func (rm *RegistryManager) mergeRegistry(imported *Registry) error {
	// Merge machines (skip duplicates)
	for _, importedMachine := range imported.Machines {
		exists := false
		for _, existingMachine := range rm.registry.Machines {
			if existingMachine.Name == importedMachine.Name {
				exists = true
				break
			}
		}
		if !exists {
			rm.registry.Machines = append(rm.registry.Machines, importedMachine)
			log.Info(fmt.Sprintf("Imported machine '%s'", importedMachine.Name))
		} else {
			log.Warn(fmt.Sprintf("Skipped duplicate machine '%s' during import", importedMachine.Name))
		}
	}

	// Merge groups (skip duplicates)
	for _, importedGroup := range imported.Groups {
		exists := false
		for _, existingGroup := range rm.registry.Groups {
			if existingGroup.Name == importedGroup.Name {
				exists = true
				break
			}
		}
		if !exists {
			rm.registry.Groups = append(rm.registry.Groups, importedGroup)
			log.Info(fmt.Sprintf("Imported group '%s'", importedGroup.Name))
		} else {
			log.Warn(fmt.Sprintf("Skipped duplicate group '%s' during import", importedGroup.Name))
		}
	}

	return rm.saveRegistry()
}
