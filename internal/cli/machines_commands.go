package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// createMachinesCommand creates the machines command with all subcommands
func createMachinesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machines",
		Short: "Manage and deploy NixOS configurations across multiple machines (flake.nix-based)",
		Long: `Multi-Machine Configuration Manager for NixOS (flake.nix-based).

Manage and deploy NixOS configurations across multiple machines using flake.nix as the single source of truth.

Features:
- List all hosts from flake.nix nixosConfigurations
- Deploy configurations remotely using native NixOS tools
- No registry or custom YAML files required

Examples:
  nixai machines list                    # List all hosts from flake.nix
  nixai machines deploy --machine myhost # Deploy to a specific host
  nixai machines deploy --method deploy-rs # Use deploy-rs if configured
`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	cmd.AddCommand(createMachinesListCommand())
	cmd.AddCommand(createMachinesDeployCommand())
	cmd.AddCommand(createMachinesSetupDeployRsCommand())
	return cmd
}

// createMachinesListCommand creates the machines list command
func createMachinesListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all NixOS hosts from flake.nix",
		Long: `List all NixOS hosts defined in nixosConfigurations in flake.nix.

This command enumerates all hosts from your flake.nix, which is now the single source of truth for machine management.

Examples:
  nixai machines list
  nixai machines list --format json`,
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			debug := os.Getenv("NIXAI_DEBUG") == "1"
			hosts, err := utils.GetFlakeHosts("", debug)
			if err != nil {
				fmt.Printf("Error: Failed to enumerate hosts from flake.nix: %v\n", err)
				os.Exit(1)
			}
			if len(hosts) == 0 {
				fmt.Println(utils.FormatInfo("No hosts found in flake.nix nixosConfigurations."))
				return
			}
			switch format {
			case "json":
				data, _ := json.MarshalIndent(hosts, "", "  ")
				fmt.Println(string(data))
			default:
				fmt.Println(utils.FormatHeader("NixOS Hosts from flake.nix:"))
				for _, h := range hosts {
					fmt.Println("-", h)
				}
			}
		},
	}
	cmd.Flags().String("format", "", "Output format: table (default) or json")
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
			fmt.Println(utils.FormatHeader("🚀 Configuration Deployment"))
			fmt.Println()

			method, _ := cmd.Flags().GetString("method")
			machine := cmd.Flag("machine").Value.String()
			group := cmd.Flag("group").Value.String()
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			if method == "deploy-rs" {
				// Check if deploy-rs is installed
				if _, err := utils.LookPath("deploy"); err != nil {
					fmt.Println(utils.FormatError("deploy-rs is not installed. Install it with: nix profile install nixpkgs#deploy-rs"))
					return
				}
				// Check for deploy config in flake.nix
				flakeDir := "."
				if nixosPath := cmd.Flag("nixos-path").Value.String(); nixosPath != "" {
					flakeDir = nixosPath
				}

				flakePath := flakeDir + "/flake.nix"
				if !utils.FlakeHasDeployConfig(flakePath) {
					fmt.Println(utils.FormatWarning("No deploy-rs configuration found in flake.nix."))
					fmt.Println(utils.FormatInfo("nixai can automatically generate a deploy-rs configuration based on your existing nixosConfigurations."))
					fmt.Println()

					if utils.PromptYesNo("Generate deploy-rs configuration for all your hosts?") {
						fmt.Println(utils.FormatInfo("Generating deploy-rs configuration..."))
						fmt.Println(utils.FormatInfo("This will add deploy-rs input and configuration to your flake.nix"))
						fmt.Println()

						// Use interactive mode to prompt for hostnames and SSH users
						if err := utils.GenerateDeployRsConfig(flakeDir, true); err != nil {
							fmt.Println(utils.FormatError("Failed to generate deploy-rs config: " + err.Error()))
							return
						}

						fmt.Println()
						fmt.Println(utils.FormatSuccess("✅ Deploy-rs configuration generated successfully!"))
						fmt.Println(utils.FormatInfo("📝 Please review the generated configuration in flake.nix"))
						fmt.Println(utils.FormatInfo("🔧 You may need to adjust hostnames and SSH settings"))
						fmt.Println(utils.FormatInfo("📚 See: https://github.com/serokell/deploy-rs#configuration"))
						fmt.Println()
					} else {
						fmt.Println(utils.FormatInfo("Aborting deploy. Please add deploy-rs config manually and try again."))
						fmt.Println(utils.FormatInfo("Or run: nixai machines setup-deploy-rs"))
						return
					}
				}
				// Run deploy-rs
				cmdArgs := []string{"deploy", "--auto-rollback", "true"}
				if group != "" {
					cmdArgs = append(cmdArgs, "--group", group)
				}
				if machine != "" {
					cmdArgs = append(cmdArgs, "--hostname", machine)
				}
				if dryRun {
					cmdArgs = append(cmdArgs, "--dry-activate")
				}
				fmt.Println(utils.FormatInfo("Running deploy-rs: deploy " + strings.Join(cmdArgs[1:], " ")))
				if err := utils.RunCommand("deploy", cmdArgs[1:]...); err != nil {
					fmt.Println(utils.FormatError("deploy-rs failed: " + err.Error()))
					return
				}
				fmt.Println(utils.FormatSuccess("deploy-rs deployment complete."))
				return
			}
			// Default: flakes (nixos-rebuild)
			fmt.Println(utils.FormatInfo("Using flakes (nixos-rebuild) deployment method."))
			// ...existing code for flakes deployment...
		},
	}

	cmd.Flags().String("machine", "", "Deploy to specific machine")
	cmd.Flags().String("group", "", "Deploy to all machines in group")
	cmd.Flags().Bool("dry-run", false, "Show what would be deployed without making changes")
	cmd.Flags().String("method", "flakes", "Deployment method: flakes (default) or deploy-rs")

	return cmd
}

// createMachinesSetupDeployRsCommand creates the setup-deploy-rs command
func createMachinesSetupDeployRsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-deploy-rs",
		Short: "Setup deploy-rs configuration for your flake.nix",
		Long: `Setup deploy-rs configuration based on your existing nixosConfigurations.

This command will:
1. Add deploy-rs input to your flake.nix if not present
2. Generate deploy configuration for all hosts in nixosConfigurations
3. Prompt for hostnames and SSH users for each host
4. Create a complete deploy-rs configuration

Examples:
  nixai machines setup-deploy-rs               # Interactive setup
  nixai machines setup-deploy-rs --non-interactive # Use defaults`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(utils.FormatHeader("🚀 Deploy-rs Configuration Setup"))
			fmt.Println()

			// Get flake directory
			flakeDir := "."
			if nixosPath := cmd.Flag("nixos-path").Value.String(); nixosPath != "" {
				flakeDir = nixosPath
			}

			interactive := !cmd.Flag("non-interactive").Changed

			// Check if flake.nix exists
			flakePath := flakeDir + "/flake.nix"
			if !utils.IsFile(flakePath) {
				fmt.Println(utils.FormatError("flake.nix not found in " + flakeDir))
				fmt.Println(utils.FormatInfo("Please run this command from your NixOS configuration directory"))
				return
			}

			// Check if deploy-rs config already exists
			if utils.FlakeHasDeployConfig(flakePath) {
				fmt.Println(utils.FormatWarning("Deploy-rs configuration already exists in flake.nix"))
				if !utils.PromptYesNo("Overwrite existing configuration?") {
					fmt.Println(utils.FormatInfo("Setup cancelled."))
					return
				}
			}

			// Get existing hosts
			hosts, err := utils.GetFlakeHosts(flakeDir)
			if err != nil {
				fmt.Println(utils.FormatError("Failed to read hosts from flake.nix: " + err.Error()))
				return
			}

			if len(hosts) == 0 {
				fmt.Println(utils.FormatError("No hosts found in nixosConfigurations"))
				fmt.Println(utils.FormatInfo("Please add some nixosConfigurations to your flake.nix first"))
				return
			}

			fmt.Println(utils.FormatInfo("Found hosts: " + strings.Join(hosts, ", ")))
			fmt.Println()

			if interactive {
				fmt.Println(utils.FormatInfo("Setting up deploy-rs configuration..."))
				fmt.Println(utils.FormatInfo("For each host, you'll be prompted for:"))
				fmt.Println("  • Hostname/IP address (defaults to host name)")
				fmt.Println("  • SSH user (defaults to 'nixos')")
				fmt.Println()
			}

			// Generate deploy-rs configuration
			if err := utils.GenerateDeployRsConfig(flakeDir, interactive); err != nil {
				fmt.Println(utils.FormatError("Failed to generate deploy-rs config: " + err.Error()))
				return
			}

			fmt.Println()
			fmt.Println(utils.FormatSuccess("✅ Deploy-rs configuration generated successfully!"))
			fmt.Println()
			fmt.Println(utils.FormatInfo("📝 Next steps:"))
			fmt.Println("  1. Review the generated configuration in flake.nix")
			fmt.Println("  2. Update hostnames and SSH settings as needed")
			fmt.Println("  3. Install deploy-rs: nix profile install nixpkgs#deploy-rs")
			fmt.Println("  4. Test deployment: nixai machines deploy --method deploy-rs --dry-run")
			fmt.Println()
			fmt.Println(utils.FormatInfo("📚 Documentation: https://github.com/serokell/deploy-rs#configuration"))
		},
	}

	cmd.Flags().Bool("non-interactive", false, "Use default values without prompting")
	return cmd
}
