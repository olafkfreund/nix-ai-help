package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/internal/machines"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// createMachinesCommand creates the machines command with all subcommands
func createMachinesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machines",
		Short: "Manage and synchronize NixOS configurations across multiple machines",
		Long: `Multi-Machine Configuration Manager for NixOS.

Manage and synchronize NixOS configurations across multiple machines with:
- Machine registry for centralized management
- Configuration synchronization between machines
- Remote deployment with rollback support
- Configuration comparison and diff analysis
- Fleet management with machine groups

Features:
- Register and manage multiple NixOS machines
- Sync configurations between local and remote machines
- Deploy configurations remotely with safety checks
- Compare configurations across machines
- Group machines for fleet operations
- SSH-based secure communication
- Automatic status monitoring
- Rollback capabilities

Examples:
  nixai machines list                    # List all registered machines
  nixai machines add server1 192.168.1.10  # Register new machine
  nixai machines sync server1           # Sync configs to machine
  nixai machines diff                   # Compare configurations
  nixai machines deploy --group webservers  # Deploy to machine group`,
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no subcommand provided
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(createMachinesListCommand())
	cmd.AddCommand(createMachinesAddCommand())
	cmd.AddCommand(createMachinesRemoveCommand())
	cmd.AddCommand(createMachinesShowCommand())
	cmd.AddCommand(createMachinesUpdateCommand())
	cmd.AddCommand(createMachinesSyncCommand())
	cmd.AddCommand(createMachinesDeployCommand())
	cmd.AddCommand(createMachinesDiffCommand())
	cmd.AddCommand(createMachinesStatusCommand())
	cmd.AddCommand(createMachinesGroupsCommand())

	return cmd
}

// createMachinesListCommand creates the machines list command
func createMachinesListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered machines",
		Long: `List all registered machines in the registry.

Displays machine information including:
- Name and host address
- Connection status
- Last sync and deploy times
- Tags and groups
- Machine metadata

Output can be formatted as table, JSON, or YAML.`,
		Example: `  nixai machines list
  nixai machines list --tag web
  nixai machines list --group production
  nixai machines list --format json`,
		Run: func(cmd *cobra.Command, args []string) {
			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			// Get filter options
			tag, _ := cmd.Flags().GetString("tag")
			group, _ := cmd.Flags().GetString("group")
			format, _ := cmd.Flags().GetString("format")
			status, _ := cmd.Flags().GetString("status")

			// Get machines list
			var machinesList []machines.Machine
			if group != "" {
				groupMachines, err := registry.GetMachinesByGroup(group)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				machinesList = groupMachines
			} else if tag != "" {
				machinesList = registry.GetMachinesByTag(tag)
			} else {
				machinesList = registry.ListMachines()
			}

			// Filter by status if specified
			if status != "" {
				var filtered []machines.Machine
				targetStatus := machines.MachineStatus(status)
				for _, machine := range machinesList {
					if machine.Status == targetStatus {
						filtered = append(filtered, machine)
					}
				}
				machinesList = filtered
			}

			if len(machinesList) == 0 {
				fmt.Println(utils.FormatInfo(fmt.Sprintf("No machines found matching the criteria")))
				return
			}

			// Format output
			switch format {
			case "json":
				data, _ := json.MarshalIndent(machinesList, "", "  ")
				fmt.Println(string(data))
			case "yaml":
				// TODO: Add YAML formatting if needed
				fmt.Println("YAML format not yet implemented")
			default:
				displayMachinesTable(machinesList)
			}
		},
	}

	cmd.Flags().String("tag", "", "Filter machines by tag")
	cmd.Flags().String("group", "", "Filter machines by group")
	cmd.Flags().String("status", "", "Filter machines by status (online, offline, syncing, deploying, error, unknown)")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")

	return cmd
}

// createMachinesAddCommand creates the machines add command
func createMachinesAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <host>",
		Short: "Register a new machine in the registry",
		Long: `Register a new NixOS machine for management.

The machine will be added to the registry with the specified name and host.
You can optionally specify additional connection details and metadata.

Examples:
  nixai machines add server1 192.168.1.10
  nixai machines add server1 server1.example.com --user nixos --port 2222
  nixai machines add server1 192.168.1.10 --nixos-path /etc/nixos --tag production`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			host := args[1]

			// Get optional flags
			user, _ := cmd.Flags().GetString("user")
			port, _ := cmd.Flags().GetInt("port")
			sshKey, _ := cmd.Flags().GetString("ssh-key")
			nixosPath, _ := cmd.Flags().GetString("nixos-path")
			description, _ := cmd.Flags().GetString("description")
			tags, _ := cmd.Flags().GetStringSlice("tag")

			// Create machine object
			machine := machines.Machine{
				Name:        name,
				Host:        host,
				Port:        port,
				User:        user,
				SSHKey:      sshKey,
				NixOSPath:   nixosPath,
				Description: description,
				Tags:        tags,
				Metadata:    make(map[string]string),
			}

			// Set default values
			if machine.Port == 0 {
				machine.Port = 22
			}
			if machine.User == "" {
				machine.User = "root"
			}
			if machine.NixOSPath == "" {
				machine.NixOSPath = "/etc/nixos"
			}

			// Add to registry
			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			if err := registry.AddMachine(machine); err != nil {
				fmt.Printf("Error: Failed to add machine: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatHeader(fmt.Sprintf("🖥️  Machine Added Successfully")))
			fmt.Println()
			fmt.Println(utils.FormatKeyValue("Name", machine.Name))
			fmt.Println(utils.FormatKeyValue("Host", machine.Host))
			fmt.Println(utils.FormatKeyValue("Port", strconv.Itoa(machine.Port)))
			fmt.Println(utils.FormatKeyValue("User", machine.User))
			fmt.Println(utils.FormatKeyValue("NixOS Path", machine.NixOSPath))
			if machine.Description != "" {
				fmt.Println(utils.FormatKeyValue("Description", machine.Description))
			}
			if len(machine.Tags) > 0 {
				fmt.Println(utils.FormatKeyValue("Tags", strings.Join(machine.Tags, ", ")))
			}
			fmt.Println()
			fmt.Println(utils.FormatInfo(fmt.Sprintf("Machine '%s' has been registered. Use 'nixai machines status %s' to test connectivity.", machine.Name, machine.Name)))
		},
	}

	cmd.Flags().String("user", "root", "SSH user (default: root)")
	cmd.Flags().Int("port", 22, "SSH port (default: 22)")
	cmd.Flags().String("ssh-key", "", "Path to SSH private key")
	cmd.Flags().String("nixos-path", "/etc/nixos", "Path to NixOS configuration on remote machine")
	cmd.Flags().String("description", "", "Description of the machine")
	cmd.Flags().StringSlice("tag", []string{}, "Tags to apply to the machine (can be used multiple times)")

	return cmd
}

// createMachinesRemoveCommand creates the machines remove command
func createMachinesRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a machine from the registry",
		Long: `Remove a machine from the registry.

This will permanently remove the machine from the registry and from all groups.
The machine's configuration files will not be affected.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			// Check if machine exists
			machine, err := registry.GetMachine(name)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Confirm removal unless --force is used
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				fmt.Printf("Are you sure you want to remove machine '%s' (%s)? [y/N]: ", machine.Name, machine.Host)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Removal cancelled.")
					return
				}
			}

			if err := registry.RemoveMachine(name); err != nil {
				fmt.Printf("Error: Failed to remove machine: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Machine '%s' removed from registry", name)))
		},
	}

	cmd.Flags().Bool("force", false, "Remove without confirmation")

	return cmd
}

// createMachinesShowCommand creates the machines show command
func createMachinesShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show detailed information about a machine",
		Long:  `Display detailed information about a registered machine.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			machine, err := registry.GetMachine(name)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatHeader(fmt.Sprintf("🖥️  Machine Details: " + machine.Name)))
			fmt.Println()
			fmt.Println(utils.FormatKeyValue("Name", machine.Name))
			fmt.Println(utils.FormatKeyValue("Host", machine.Host))
			fmt.Println(utils.FormatKeyValue("Port", strconv.Itoa(machine.Port)))
			fmt.Println(utils.FormatKeyValue("User", machine.User))
			fmt.Println(utils.FormatKeyValue("NixOS Path", machine.NixOSPath))
			fmt.Println(utils.FormatKeyValue("Status", string(machine.Status)))

			if machine.Description != "" {
				fmt.Println(utils.FormatKeyValue("Description", machine.Description))
			}

			if len(machine.Tags) > 0 {
				fmt.Println(utils.FormatKeyValue("Tags", strings.Join(machine.Tags, ", ")))
			}

			if len(machine.Groups) > 0 {
				fmt.Println(utils.FormatKeyValue("Groups", strings.Join(machine.Groups, ", ")))
			}

			if machine.SSHKey != "" {
				fmt.Println(utils.FormatKeyValue("SSH Key", machine.SSHKey))
			}

			fmt.Println(utils.FormatKeyValue("Created", machine.CreatedAt.Format(time.RFC3339)))
			fmt.Println(utils.FormatKeyValue("Updated", machine.UpdatedAt.Format(time.RFC3339)))

			if machine.LastSync != nil {
				fmt.Println(utils.FormatKeyValue("Last Sync", machine.LastSync.Format(time.RFC3339)))
			}

			if machine.LastDeploy != nil {
				fmt.Println(utils.FormatKeyValue("Last Deploy", machine.LastDeploy.Format(time.RFC3339)))
			}

			// Show metadata if any
			if len(machine.Metadata) > 0 {
				fmt.Println()
				fmt.Println(utils.FormatSubheader(fmt.Sprintf("Metadata:")))
				for key, value := range machine.Metadata {
					fmt.Println(utils.FormatKeyValue(key, value))
				}
			}
		},
	}

	return cmd
}

// createMachinesUpdateCommand creates the machines update command
func createMachinesUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update machine configuration",
		Long:  `Update the configuration of a registered machine.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			machine, err := registry.GetMachine(name)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Update fields if flags are provided
			if cmd.Flags().Changed("host") {
				host, _ := cmd.Flags().GetString("host")
				machine.Host = host
			}
			if cmd.Flags().Changed("port") {
				port, _ := cmd.Flags().GetInt("port")
				machine.Port = port
			}
			if cmd.Flags().Changed("user") {
				user, _ := cmd.Flags().GetString("user")
				machine.User = user
			}
			if cmd.Flags().Changed("ssh-key") {
				sshKey, _ := cmd.Flags().GetString("ssh-key")
				machine.SSHKey = sshKey
			}
			if cmd.Flags().Changed("nixos-path") {
				nixosPath, _ := cmd.Flags().GetString("nixos-path")
				machine.NixOSPath = nixosPath
			}
			if cmd.Flags().Changed("description") {
				description, _ := cmd.Flags().GetString("description")
				machine.Description = description
			}
			if cmd.Flags().Changed("tag") {
				tags, _ := cmd.Flags().GetStringSlice("tag")
				machine.Tags = tags
			}

			if err := registry.UpdateMachine(name, *machine); err != nil {
				fmt.Printf("Error: Failed to update machine: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Machine '%s' updated successfully", name)))
		},
	}

	cmd.Flags().String("host", "", "Update host address")
	cmd.Flags().Int("port", 0, "Update SSH port")
	cmd.Flags().String("user", "", "Update SSH user")
	cmd.Flags().String("ssh-key", "", "Update SSH key path")
	cmd.Flags().String("nixos-path", "", "Update NixOS configuration path")
	cmd.Flags().String("description", "", "Update description")
	cmd.Flags().StringSlice("tag", []string{}, "Update tags (replaces existing tags)")

	return cmd
}

// createMachinesSyncCommand creates the machines sync command
func createMachinesSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync <machine>",
		Short: "Synchronize configurations to a machine",
		Long: `Synchronize NixOS configurations from local machine to remote machine.

This command will:
1. Copy configuration files to the remote machine
2. Verify file integrity
3. Update sync timestamps
4. Provide rollback information if needed`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			machineName := args[0]

			// TODO: Implement sync functionality
			fmt.Println(utils.FormatHeader(fmt.Sprintf("🔄 Configuration Sync")))
			fmt.Println()
			fmt.Printf("Synchronizing configurations to machine '%s'...\n", machineName)
			fmt.Println()
			fmt.Println(utils.FormatInfo(fmt.Sprintf("Sync functionality will be implemented in the next iteration.")))
			fmt.Println(utils.FormatInfo(fmt.Sprintf("This will include:")))
			fmt.Println("  • SSH-based file transfer with rsync")
			fmt.Println("  • Configuration validation")
			fmt.Println("  • Integrity verification")
			fmt.Println("  • Automatic rollback on failure")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be synced without making changes")
	cmd.Flags().String("source", "", "Source directory to sync (default: current nixos-path)")
	cmd.Flags().Bool("force", false, "Force sync even if target is newer")

	return cmd
}

// createMachinesDeployCommand creates the machines deploy command
func createMachinesDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy configurations to machines",
		Long: `Deploy NixOS configurations to one or more machines.

This command will:
1. Sync configurations if needed
2. Run nixos-rebuild on remote machines  
3. Monitor deployment progress
4. Provide rollback commands if deployment fails`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement deploy functionality
			fmt.Println(utils.FormatHeader(fmt.Sprintf("🚀 Configuration Deployment")))
			fmt.Println()
			fmt.Println(utils.FormatInfo(fmt.Sprintf("Deploy functionality will be implemented in the next iteration.")))
			fmt.Println(utils.FormatInfo(fmt.Sprintf("This will include:")))
			fmt.Println("  • Remote nixos-rebuild execution")
			fmt.Println("  • Generation management")
			fmt.Println("  • Parallel deployment to multiple machines")
			fmt.Println("  • Automatic rollback on failure")
			fmt.Println("  • Deployment progress monitoring")
		},
	}

	cmd.Flags().String("machine", "", "Deploy to specific machine")
	cmd.Flags().String("group", "", "Deploy to all machines in group")
	cmd.Flags().String("tag", "", "Deploy to all machines with tag")
	cmd.Flags().Bool("dry-run", false, "Show what would be deployed without making changes")
	cmd.Flags().String("action", "switch", "Deployment action (switch, boot, test)")

	return cmd
}

// createMachinesDiffCommand creates the machines diff command
func createMachinesDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [machine1] [machine2]",
		Short: "Compare configurations between machines",
		Long: `Compare NixOS configurations between machines.

If no machines are specified, compares all machines.
If one machine is specified, compares it with the local configuration.
If two machines are specified, compares them with each other.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement diff functionality
			fmt.Println(utils.FormatHeader(fmt.Sprintf("🔍 Configuration Comparison")))
			fmt.Println()

			switch len(args) {
			case 0:
				fmt.Println(utils.FormatInfo(fmt.Sprintf("Comparing all registered machines...")))
			case 1:
				fmt.Printf("Comparing machine '%s' with local configuration...\n", args[0])
			case 2:
				fmt.Printf("Comparing machines '%s' and '%s'...\n", args[0], args[1])
			default:
				fmt.Println("Error: Too many arguments. Expected 0-2 machine names.")
				os.Exit(1)
			}

			fmt.Println()
			fmt.Println(utils.FormatInfo(fmt.Sprintf("Diff functionality will be implemented in the next iteration.")))
			fmt.Println(utils.FormatInfo(fmt.Sprintf("This will include:")))
			fmt.Println("  • Configuration file comparison")
			fmt.Println("  • Package difference analysis")
			fmt.Println("  • Service configuration comparison")
			fmt.Println("  • AI-powered difference explanations")
		},
	}

	cmd.Flags().Bool("summary", false, "Show only summary of differences")
	cmd.Flags().String("format", "unified", "Diff format (unified, context, side-by-side)")

	return cmd
}

// createMachinesStatusCommand creates the machines status command
func createMachinesStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [machine]",
		Short: "Check status and connectivity of machines",
		Long: `Check the status and connectivity of one or all machines.

This command will:
1. Test SSH connectivity
2. Check NixOS system status  
3. Update machine status in registry
4. Show system information if available`,
		Run: func(cmd *cobra.Command, args []string) {
			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			var machinesToCheck []machines.Machine

			if len(args) == 1 {
				// Check specific machine
				machine, err := registry.GetMachine(args[0])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				machinesToCheck = []machines.Machine{*machine}
			} else {
				// Check all machines
				machinesToCheck = registry.ListMachines()
			}

			if len(machinesToCheck) == 0 {
				fmt.Println(utils.FormatInfo(fmt.Sprintf("No machines to check")))
				return
			}

			fmt.Println(utils.FormatHeader(fmt.Sprintf("🔍 Machine Status Check")))
			fmt.Println()

			// TODO: Implement actual connectivity testing
			for _, machine := range machinesToCheck {
				fmt.Printf("Checking %s (%s)... ", machine.Name, machine.Host)

				// Simulate status check - will be replaced with actual SSH connectivity test
				fmt.Println(utils.FormatInfo(fmt.Sprintf("Status check will be implemented in next iteration")))

				// For now, just show current status
				fmt.Println(utils.FormatKeyValue("  Current Status", string(machine.Status)))
				fmt.Println()
			}

			fmt.Println(utils.FormatInfo(fmt.Sprintf("Status check functionality will include:")))
			fmt.Println("  • SSH connectivity testing")
			fmt.Println("  • NixOS system status verification")
			fmt.Println("  • Automatic status updates in registry")
			fmt.Println("  • System information gathering")
		},
	}

	return cmd
}

// createMachinesGroupsCommand creates the machines groups command
func createMachinesGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage machine groups",
		Long:  `Manage groups of machines for fleet operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add group subcommands
	cmd.AddCommand(createGroupsListCommand())
	cmd.AddCommand(createGroupsCreateCommand())
	cmd.AddCommand(createGroupsDeleteCommand())
	cmd.AddCommand(createGroupsAddMachineCommand())
	cmd.AddCommand(createGroupsRemoveMachineCommand())

	return cmd
}

// Helper function to display machines in a table format
func displayMachinesTable(machinesList []machines.Machine) {
	if len(machinesList) == 0 {
		fmt.Println(utils.FormatInfo(fmt.Sprintf("No machines found")))
		return
	}

	fmt.Println(utils.FormatHeader(fmt.Sprintf("🖥️  Registered Machines")))
	fmt.Println()

	// Sort by name for consistent output
	sort.Slice(machinesList, func(i, j int) bool {
		return machinesList[i].Name < machinesList[j].Name
	})

	// Print table header
	fmt.Printf("%-20s %-20s %-10s %-15s %-20s %-15s\n",
		"NAME", "HOST", "STATUS", "LAST SYNC", "LAST DEPLOY", "TAGS")
	fmt.Println(strings.Repeat("-", 100))

	// Print machine rows
	for _, machine := range machinesList {
		lastSync := "Never"
		if machine.LastSync != nil {
			lastSync = machine.LastSync.Format("2006-01-02 15:04")
		}

		lastDeploy := "Never"
		if machine.LastDeploy != nil {
			lastDeploy = machine.LastDeploy.Format("2006-01-02 15:04")
		}

		tags := strings.Join(machine.Tags, ",")
		if len(tags) > 14 {
			tags = tags[:11] + "..."
		}

		fmt.Printf("%-20s %-20s %-10s %-15s %-20s %-15s\n",
			truncateString(machine.Name, 19),
			truncateString(machine.Host, 19),
			string(machine.Status),
			lastSync,
			lastDeploy,
			tags)
	}

	fmt.Println()
	fmt.Printf("Total: %d machines\n", len(machinesList))
}

// Helper function to truncate strings for table display
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Group management subcommands
func createGroupsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List machine groups",
		Run: func(cmd *cobra.Command, args []string) {
			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			groups := registry.ListGroups()
			if len(groups) == 0 {
				fmt.Println(utils.FormatInfo(fmt.Sprintf("No machine groups found")))
				return
			}

			fmt.Println(utils.FormatHeader(fmt.Sprintf("👥 Machine Groups")))
			fmt.Println()

			for _, group := range groups {
				fmt.Printf("📁 %s (%d machines)\n", group.Name, len(group.Machines))
				if group.Description != "" {
					fmt.Printf("   %s\n", group.Description)
				}
				fmt.Printf("   Machines: %s\n", strings.Join(group.Machines, ", "))
				fmt.Println()
			}
		},
	}
}

func createGroupsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new machine group",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			group := machines.MachineGroup{
				Name:        name,
				Description: description,
				Machines:    []string{},
			}

			if err := registry.AddGroup(group); err != nil {
				fmt.Printf("Error: Failed to create group: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Group '%s' created successfully", name)))
		},
	}
}

func createGroupsDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a machine group",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			if err := registry.RemoveGroup(name); err != nil {
				fmt.Printf("Error: Failed to delete group: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Group '%s' deleted successfully", name)))
		},
	}
}

func createGroupsAddMachineCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-machine <group> <machine>",
		Short: "Add a machine to a group",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			machineName := args[1]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			if err := registry.AddMachineToGroup(groupName, machineName); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Machine '%s' added to group '%s'", machineName, groupName)))
		},
	}
}

func createGroupsRemoveMachineCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-machine <group> <machine>",
		Short: "Remove a machine from a group",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			groupName := args[0]
			machineName := args[1]

			registry, err := machines.NewRegistryManager()
			if err != nil {
				fmt.Printf("Error: Failed to load machine registry: %v\n", err)
				os.Exit(1)
			}

			if err := registry.RemoveMachineFromGroup(groupName, machineName); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(utils.FormatSuccess(fmt.Sprintf("Machine '%s' removed from group '%s'", machineName, groupName)))
		},
	}
}
