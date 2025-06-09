package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/devenv"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// devenvCmd is the main devenv command
var devenvCmd = &cobra.Command{
	Use:   "devenv",
	Short: "Create and manage development environments with devenv",
	Long: `Create and manage development environments using devenv templates.

devenv is a tool for creating reproducible development environments using Nix.
This command helps you quickly set up development environments for different
programming languages and frameworks.

Examples:
  nixai devenv list                    # List available templates
  nixai devenv create python myproject # Create Python environment
  nixai devenv create rust --with-wasm # Create Rust environment with WebAssembly
  nixai devenv suggest "web app with database" # Get AI template suggestion`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// NewDevenvListCmd returns a fresh list command
func NewDevenvListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available devenv templates",
		Long:  "List all available devenv templates with their descriptions.", Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
				os.Exit(1)
			}
			log := logger.NewLoggerWithLevel(cfg.LogLevel)

			// Use the new ProviderManager system
			aiProvider, err := GetLegacyAIProvider(cfg, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error getting AI provider: "+err.Error()))
				os.Exit(1)
			}

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error creating devenv service: "+err.Error()))
				os.Exit(1)
			}
			fmt.Println(utils.FormatHeader("📦 Available Development Environment Templates"))
			fmt.Println(utils.FormatDivider())
			templates := service.ListTemplates()
			if len(templates) == 0 {
				fmt.Println(utils.FormatWarning("No templates available"))
				return
			}
			names := make([]string, 0, len(templates))
			for name := range templates {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				description := templates[name]
				status := "✅ Enabled"
				if templateConfig, exists := cfg.Devenv.Templates[name]; exists && !templateConfig.Enabled {
					status = "❌ Disabled"
				}
				fmt.Printf("  %s %s\n", utils.FormatKeyValue(name, description), utils.FormatNote(status))
			}
			fmt.Println("\n" + utils.FormatTip("Use 'nixai devenv create <template> <project-name>' to create a new project"))
			fmt.Println(utils.FormatTip("Use 'nixai devenv suggest \"<description>\"' for AI-powered template suggestions"))
		},
	}
}

// NewDevenvCreateCmd returns a fresh create command
func NewDevenvCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <template> [project-name]",
		Short: "Create a new development environment from a template",
		Long:  "Create a new development environment using a specific template.\n\nThe command will create a new directory with the project name (if specified)\nor use the current directory, and generate a devenv.nix file along with\ntemplate-specific starter files.\n\nExamples:\n  nixai devenv create python myapp\n  nixai devenv create rust --with-wasm --services postgres\n  nixai devenv create nodejs --framework nextjs --directory ./my-web-app\n  nixai devenv create golang --with-grpc --services redis,postgres",
		Args:  conditionalRangeArgsValidator(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			templateName := args[0]
			projectName := ""
			if len(args) > 1 {
				projectName = args[1]
			} else {
				cwd, _ := os.Getwd()
				projectName = filepath.Base(cwd)
			}
			directory, _ := cmd.Flags().GetString("directory")
			servicesFlag, _ := cmd.Flags().GetString("services")
			interactive, _ := cmd.Flags().GetBool("interactive")
			var services []string
			if servicesFlag != "" {
				services = strings.Split(servicesFlag, ",")
				for i, service := range services {
					services[i] = strings.TrimSpace(service)
				}
			}
			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
				os.Exit(1)
			}
			log := logger.NewLoggerWithLevel(cfg.LogLevel)

			// Use the new ProviderManager system
			aiProvider, err := GetLegacyAIProvider(cfg, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error getting AI provider: "+err.Error()))
				os.Exit(1)
			}

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error creating devenv service: "+err.Error()))
				os.Exit(1)
			}
			template, err := service.GetTemplate(templateName)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Template not found: "+templateName))
				fmt.Println(utils.FormatInfo("Use 'nixai devenv list' to see available templates"))
				os.Exit(1)
			}
			fmt.Println(utils.FormatHeader("🚀 Creating Development Environment"))
			fmt.Println(utils.FormatKeyValue("Template", templateName))
			fmt.Println(utils.FormatKeyValue("Project", projectName))
			if directory != "" {
				fmt.Println(utils.FormatKeyValue("Directory", directory))
			}
			if len(services) > 0 {
				fmt.Println(utils.FormatKeyValue("Services", strings.Join(services, ", ")))
			}
			fmt.Println(utils.FormatDivider())
			var options map[string]string
			if interactive {
				options = collectTemplateOptions(template)
			} else {
				options = collectFlagOptions(cmd, template, cfg)
			}
			fmt.Println(utils.FormatProgress("Generating devenv configuration..."))
			err = service.CreateProject(templateName, projectName, directory, options, services)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error creating project: "+err.Error()))
				os.Exit(1)
			}
			finalDir := directory
			if finalDir == "" {
				finalDir = projectName
			}
			absDir, _ := filepath.Abs(finalDir)
			fmt.Println(utils.FormatSuccess("✅ Development environment created successfully!"))
			fmt.Println()
			fmt.Println(utils.FormatKeyValue("Location", absDir))
			fmt.Println(utils.FormatKeyValue("devenv.nix", filepath.Join(absDir, "devenv.nix")))
			fmt.Println()
			fmt.Println(utils.FormatHeader("Next Steps:"))
			fmt.Printf("  1. %s\n", utils.FormatNote("cd "+finalDir))
			fmt.Printf("  2. %s\n", utils.FormatNote("devenv shell  # Enter the development environment"))
			fmt.Printf("  3. %s\n", utils.FormatNote("devenv up     # Start services (if any)"))
			fmt.Println()
			fmt.Println(utils.FormatTip("Edit devenv.nix to customize your environment"))
			fmt.Println(utils.FormatTip("Use 'devenv --help' to learn more about devenv commands"))
		},
	}
	cmd.Flags().StringP("directory", "d", "", "Directory to create the project in")
	cmd.Flags().StringP("services", "s", "", "Comma-separated list of services to include (postgres,redis,mysql,mongodb)")
	cmd.Flags().BoolP("interactive", "i", false, "Interactive mode for configuring template options")
	cmd.Flags().String("framework", "", "Web framework to use (depends on template)")
	cmd.Flags().Bool("with-typescript", false, "Include TypeScript support (nodejs template)")
	cmd.Flags().Bool("with-wasm", false, "Include WebAssembly support (rust template)")
	cmd.Flags().Bool("with-grpc", false, "Include gRPC support (golang template)")
	return cmd
}

// NewDevenvSuggestCmd returns a fresh suggest command
func NewDevenvSuggestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suggest <description>",
		Short: "Get AI-powered template suggestions",
		Long:  `Use AI to suggest the most appropriate development environment template\nbased on your project description.\n\nExamples:\n  nixai devenv suggest "web application with database"\n  nixai devenv suggest "machine learning project with jupyter"\n  nixai devenv suggest "microservice in rust"\n  nixai devenv suggest "react frontend with typescript"`,
		Args:  conditionalArgsValidator(1),
		Run: func(cmd *cobra.Command, args []string) {
			description := strings.Join(args, " ")
			cfg, err := config.LoadUserConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
				os.Exit(1)
			}
			log := logger.NewLoggerWithLevel(cfg.LogLevel)

			// Use the new ProviderManager system
			aiProvider, err := GetLegacyAIProvider(cfg, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error getting AI provider: "+err.Error()))
				os.Exit(1)
			}

			service, err := devenv.NewService(aiProvider, log)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error creating devenv service: "+err.Error()))
				os.Exit(1)
			}
			fmt.Println(utils.FormatHeader("🤖 AI Template Suggestion"))
			fmt.Println(utils.FormatKeyValue("Description", description))
			fmt.Println(utils.FormatDivider())
			fmt.Print(utils.FormatProgress("Analyzing your project description..."))
			suggestion, err := service.SuggestTemplate(description)
			if err != nil {
				fmt.Fprintln(os.Stderr, utils.FormatError("Error getting suggestion: "+err.Error()))
				os.Exit(1)
			}
			fmt.Printf("\r%s\n", utils.FormatSuccess("✅ Analysis complete"))
			fmt.Println()
			fmt.Println(utils.FormatKeyValue("Recommended Template", suggestion))
			template, err := service.GetTemplate(suggestion)
			if err == nil {
				fmt.Println(utils.FormatKeyValue("Description", template.Description()))
				if len(template.SupportedServices()) > 0 {
					fmt.Println(utils.FormatKeyValue("Supported Services", strings.Join(template.SupportedServices(), ", ")))
				}
			}
			fmt.Println()
			fmt.Println(utils.FormatHeader("Create Project:"))
			fmt.Printf("  %s\n", utils.FormatNote("nixai devenv create "+suggestion+" myproject"))
			fmt.Println()
			fmt.Println(utils.FormatTip("Use --interactive flag for guided setup"))
		},
	}
}

// Helper functions

func collectTemplateOptions(template devenv.Template) map[string]string {
	options := make(map[string]string)
	inputs := template.RequiredInputs()

	if len(inputs) == 0 {
		return options
	}

	fmt.Println(utils.FormatHeader("Template Configuration"))
	fmt.Println(utils.FormatNote("Configure template options (press Enter for defaults):"))
	fmt.Println()

	for _, input := range inputs {
		prompt := fmt.Sprintf("%s (%s)", input.Description, input.Name)
		if input.Default != "" {
			prompt += fmt.Sprintf(" [%s]", input.Default)
		}
		if len(input.Choices) > 0 {
			prompt += fmt.Sprintf(" (choices: %s)", strings.Join(input.Choices, ", "))
		}
		prompt += ": "

		fmt.Print(prompt)
		var value string
		_, _ = fmt.Scanln(&value)

		if value == "" && input.Default != "" {
			value = input.Default
		}

		if value != "" {
			options[input.Name] = value
		}
	}

	return options
}

func collectFlagOptions(cmd *cobra.Command, template devenv.Template, cfg *config.UserConfig) map[string]string {
	options := make(map[string]string)
	templateName := template.Name()

	// Set defaults from config if available
	if templateConfig, exists := cfg.Devenv.Templates[templateName]; exists {
		if templateConfig.DefaultVersion != "" {
			// Map version field names for different templates
			switch templateName {
			case "python":
				options["python_version"] = templateConfig.DefaultVersion
			case "rust":
				options["rust_version"] = templateConfig.DefaultVersion
			case "nodejs":
				options["nodejs_version"] = templateConfig.DefaultVersion
			case "golang":
				options["go_version"] = templateConfig.DefaultVersion
			}
		}
		if templateConfig.DefaultPackageManager != "" {
			options["package_manager"] = templateConfig.DefaultPackageManager
		}
	}

	// Override with command flags
	if cmd.Flags().Changed("with-wasm") {
		if withWasm, _ := cmd.Flags().GetBool("with-wasm"); withWasm {
			options["with_wasm"] = "true"
		}
	}
	if cmd.Flags().Changed("with-grpc") {
		if withGrpc, _ := cmd.Flags().GetBool("with-grpc"); withGrpc {
			options["with_grpc"] = "true"
		}
	}
	if cmd.Flags().Changed("framework") {
		if framework, _ := cmd.Flags().GetString("framework"); framework != "" {
			options["framework"] = framework
		}
	}
	if cmd.Flags().Changed("with-typescript") {
		if withTS, _ := cmd.Flags().GetBool("with-typescript"); withTS {
			options["with_typescript"] = "true"
		}
	}

	return options
}

func init() {
	devenvCmd.AddCommand(NewDevenvListCmd())
	devenvCmd.AddCommand(NewDevenvCreateCmd())
	devenvCmd.AddCommand(NewDevenvSuggestCmd())
}
