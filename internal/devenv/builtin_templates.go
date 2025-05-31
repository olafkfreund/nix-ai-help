package devenv

// BuiltinTemplates provides access to built-in templates without import cycles

// PythonTemplate implements the Template interface for Python development environments
type PythonTemplate struct{}

func (p *PythonTemplate) Name() string {
	return "python"
}

func (p *PythonTemplate) Description() string {
	return "Python development environment with pip, poetry, and common tools"
}

func (p *PythonTemplate) RequiredInputs() []InputField {
	return []InputField{
		{
			Name:        "python_version",
			Type:        "choice",
			Description: "Python version to use",
			Required:    false,
			Default:     "311",
			Choices:     []string{"39", "310", "311", "312"},
		},
		{
			Name:        "package_manager",
			Type:        "choice",
			Description: "Package manager preference",
			Required:    false,
			Default:     "pip",
			Choices:     []string{"pip", "poetry", "pipenv"},
		},
		{
			Name:        "with_jupyter",
			Type:        "bool",
			Description: "Include Jupyter notebook support",
			Required:    false,
			Default:     "false",
		},
		{
			Name:        "with_django",
			Type:        "bool",
			Description: "Include Django web framework",
			Required:    false,
			Default:     "false",
		},
		{
			Name:        "with_fastapi",
			Type:        "bool",
			Description: "Include FastAPI web framework",
			Required:    false,
			Default:     "false",
		},
	}
}

func (p *PythonTemplate) SupportedServices() []string {
	return []string{"postgres", "redis", "mysql", "mongodb"}
}

func (p *PythonTemplate) Validate(config TemplateConfig) error {
	// Basic validation - could be expanded
	return nil
}

func (p *PythonTemplate) Generate(config TemplateConfig) (*DevenvConfig, error) {
	pythonVersion := config.Options["python_version"]
	if pythonVersion == "" {
		pythonVersion = "311"
	}

	packageManager := config.Options["package_manager"]
	if packageManager == "" {
		packageManager = "pip"
	}

	devenvConfig := &DevenvConfig{
		Languages: map[string]interface{}{
			"python": map[string]interface{}{
				"enable":  true,
				"version": pythonVersion,
			},
		},
		Packages: []string{
			"git",
			"curl",
		},
		Environment: map[string]string{
			"PYTHONPATH": ".",
		},
		Scripts: make(map[string]interface{}),
	}

	// Add package manager specific tools
	switch packageManager {
	case "poetry":
		devenvConfig.Packages = append(devenvConfig.Packages, "poetry")
		devenvConfig.Scripts["install"] = "poetry install"
		devenvConfig.Scripts["run"] = "poetry run python"
		devenvConfig.Scripts["shell"] = "poetry shell"
	case "pipenv":
		devenvConfig.Packages = append(devenvConfig.Packages, "pipenv")
		devenvConfig.Scripts["install"] = "pipenv install"
		devenvConfig.Scripts["run"] = "pipenv run python"
		devenvConfig.Scripts["shell"] = "pipenv shell"
	default: // pip
		devenvConfig.Scripts["install"] = "pip install -r requirements.txt"
		devenvConfig.Scripts["run"] = "python"
	}

	// Add optional tools
	if config.Options["with_jupyter"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.jupyter")
		devenvConfig.Scripts["jupyter"] = "jupyter notebook"
	}

	if config.Options["with_django"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.django")
		devenvConfig.Scripts["django-admin"] = "django-admin"
		devenvConfig.Scripts["manage"] = "python manage.py"
	}

	if config.Options["with_fastapi"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages,
			"python3Packages.fastapi",
			"python3Packages.uvicorn")
		devenvConfig.Scripts["dev"] = "uvicorn main:app --reload"
	}

	// Add development tools
	devenvConfig.Packages = append(devenvConfig.Packages,
		"python3Packages.black",
		"python3Packages.flake8",
		"python3Packages.pytest",
		"python3Packages.mypy",
	)

	devenvConfig.Scripts["format"] = "black ."
	devenvConfig.Scripts["lint"] = "flake8 ."
	devenvConfig.Scripts["test"] = "pytest"
	devenvConfig.Scripts["typecheck"] = "mypy ."

	// Add requested services
	if len(config.Services) > 0 {
		devenvConfig.Services = make(map[string]interface{})
		for _, service := range config.Services {
			switch service {
			case "postgres":
				devenvConfig.Services["postgres"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.psycopg2")
			case "redis":
				devenvConfig.Services["redis"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.redis")
			case "mysql":
				devenvConfig.Services["mysql"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.pymysql")
			case "mongodb":
				devenvConfig.Services["mongodb"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Packages = append(devenvConfig.Packages, "python3Packages.pymongo")
			}
		}
	}

	// Add custom packages
	devenvConfig.Packages = append(devenvConfig.Packages, config.Packages...)

	// Add custom environment variables
	for key, value := range config.EnvVars {
		devenvConfig.Environment[key] = value
	}

	return devenvConfig, nil
}

// RustTemplate implements the Template interface for Rust development environments
type RustTemplate struct{}

func (r *RustTemplate) Name() string {
	return "rust"
}

func (r *RustTemplate) Description() string {
	return "Rust development environment with cargo, rustfmt, clippy, and common tools"
}

func (r *RustTemplate) RequiredInputs() []InputField {
	return []InputField{
		{
			Name:        "rust_version",
			Type:        "choice",
			Description: "Rust toolchain version",
			Required:    false,
			Default:     "stable",
			Choices:     []string{"stable", "beta", "nightly"},
		},
		{
			Name:        "with_wasm",
			Type:        "bool",
			Description: "Include WebAssembly (WASM) support",
			Required:    false,
			Default:     "false",
		},
		{
			Name:        "with_diesel",
			Type:        "bool",
			Description: "Include Diesel ORM support",
			Required:    false,
			Default:     "false",
		},
		{
			Name:        "workspace",
			Type:        "bool",
			Description: "Set up as a Cargo workspace",
			Required:    false,
			Default:     "false",
		},
	}
}

func (r *RustTemplate) SupportedServices() []string {
	return []string{"postgres", "redis", "mysql"}
}

func (r *RustTemplate) Validate(config TemplateConfig) error {
	return nil
}

func (r *RustTemplate) Generate(config TemplateConfig) (*DevenvConfig, error) {
	rustVersion := config.Options["rust_version"]
	if rustVersion == "" {
		rustVersion = "stable"
	}

	devenvConfig := &DevenvConfig{
		Languages: map[string]interface{}{
			"rust": map[string]interface{}{
				"enable":  true,
				"channel": rustVersion,
			},
		},
		Packages: []string{
			"git",
			"curl",
			"pkg-config",
			"openssl",
		},
		Environment: map[string]string{
			"RUST_BACKTRACE": "1",
		},
		Scripts: map[string]interface{}{
			"build":   "cargo build",
			"run":     "cargo run",
			"test":    "cargo test",
			"check":   "cargo check",
			"fmt":     "cargo fmt",
			"clippy":  "cargo clippy",
			"clean":   "cargo clean",
			"doc":     "cargo doc --open",
			"release": "cargo build --release",
		},
	}

	// Add WebAssembly support
	if config.Options["with_wasm"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages,
			"wasm-pack",
			"binaryen",
		)
		devenvConfig.Scripts["wasm-build"] = "wasm-pack build"
		devenvConfig.Scripts["wasm-test"] = "wasm-pack test --headless --firefox"
	}

	// Add Diesel CLI for database ORM
	if config.Options["with_diesel"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages, "diesel-cli")
		devenvConfig.Scripts["diesel"] = "diesel"
		devenvConfig.Scripts["migrate"] = "diesel migration run"
		devenvConfig.Scripts["migrate-redo"] = "diesel migration redo"
	}

	// Add workspace-specific configuration
	if config.Options["workspace"] == "true" {
		devenvConfig.Scripts["build-all"] = "cargo build --workspace"
		devenvConfig.Scripts["test-all"] = "cargo test --workspace"
		devenvConfig.Scripts["check-all"] = "cargo check --workspace"
	}

	// Add requested services
	if len(config.Services) > 0 {
		devenvConfig.Services = make(map[string]interface{})
		for _, service := range config.Services {
			switch service {
			case "postgres":
				devenvConfig.Services["postgres"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "postgres://postgres@localhost/" + config.ProjectName
			case "redis":
				devenvConfig.Services["redis"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Environment["REDIS_URL"] = "redis://localhost:6379"
			case "mysql":
				devenvConfig.Services["mysql"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "mysql://root@localhost/" + config.ProjectName
			}
		}
	}

	// Add performance and development tools
	devenvConfig.Packages = append(devenvConfig.Packages,
		"rust-analyzer", // LSP server
		"cargo-edit",    // cargo add/remove commands
		"cargo-watch",   // file watching
		"cargo-audit",   // security audit
	)

	devenvConfig.Scripts["watch"] = "cargo watch -x run"
	devenvConfig.Scripts["watch-test"] = "cargo watch -x test"
	devenvConfig.Scripts["audit"] = "cargo audit"
	devenvConfig.Scripts["edit"] = "cargo edit"

	// Add custom packages
	devenvConfig.Packages = append(devenvConfig.Packages, config.Packages...)

	// Add custom environment variables
	for key, value := range config.EnvVars {
		devenvConfig.Environment[key] = value
	}

	return devenvConfig, nil
}

// NodejsTemplate implements the Template interface for Node.js development environments
type NodejsTemplate struct{}

func (n *NodejsTemplate) Name() string {
	return "nodejs"
}

func (n *NodejsTemplate) Description() string {
	return "Node.js development environment with npm/yarn/pnpm and common tools"
}

func (n *NodejsTemplate) RequiredInputs() []InputField {
	return []InputField{
		{
			Name:        "nodejs_version",
			Type:        "choice",
			Description: "Node.js version to use",
			Required:    false,
			Default:     "20",
			Choices:     []string{"16", "18", "20", "21"},
		},
		{
			Name:        "package_manager",
			Type:        "choice",
			Description: "Package manager preference",
			Required:    false,
			Default:     "npm",
			Choices:     []string{"npm", "yarn", "pnpm"},
		},
		{
			Name:        "with_typescript",
			Type:        "bool",
			Description: "Include TypeScript support",
			Required:    false,
			Default:     "true",
		},
		{
			Name:        "framework",
			Type:        "choice",
			Description: "Web framework (optional)",
			Required:    false,
			Default:     "none",
			Choices:     []string{"none", "express", "fastify", "nextjs", "nuxtjs", "react", "vue", "svelte"},
		},
	}
}

func (n *NodejsTemplate) SupportedServices() []string {
	return []string{"postgres", "redis", "mysql", "mongodb"}
}

func (n *NodejsTemplate) Validate(config TemplateConfig) error {
	return nil
}

func (n *NodejsTemplate) Generate(config TemplateConfig) (*DevenvConfig, error) {
	nodeVersion := config.Options["nodejs_version"]
	if nodeVersion == "" {
		nodeVersion = "20"
	}

	packageManager := config.Options["package_manager"]
	if packageManager == "" {
		packageManager = "npm"
	}

	devenvConfig := &DevenvConfig{
		Languages: map[string]interface{}{
			"javascript": map[string]interface{}{
				"enable": true,
				"npm": map[string]interface{}{
					"enable": true,
				},
			},
		},
		Packages: []string{
			"git",
			"curl",
			"nodejs_" + nodeVersion,
		},
		Environment: map[string]string{
			"NODE_ENV": "development",
		},
		Scripts: make(map[string]interface{}),
	}

	// Add package manager specific tools and scripts
	switch packageManager {
	case "yarn":
		devenvConfig.Packages = append(devenvConfig.Packages, "yarn")
		devenvConfig.Scripts["install"] = "yarn install"
		devenvConfig.Scripts["dev"] = "yarn dev"
		devenvConfig.Scripts["build"] = "yarn build"
		devenvConfig.Scripts["start"] = "yarn start"
		devenvConfig.Scripts["test"] = "yarn test"
		devenvConfig.Scripts["lint"] = "yarn lint"
	case "pnpm":
		devenvConfig.Packages = append(devenvConfig.Packages, "nodePackages.pnpm")
		devenvConfig.Scripts["install"] = "pnpm install"
		devenvConfig.Scripts["dev"] = "pnpm dev"
		devenvConfig.Scripts["build"] = "pnpm build"
		devenvConfig.Scripts["start"] = "pnpm start"
		devenvConfig.Scripts["test"] = "pnpm test"
		devenvConfig.Scripts["lint"] = "pnpm lint"
	default: // npm
		devenvConfig.Scripts["install"] = "npm install"
		devenvConfig.Scripts["dev"] = "npm run dev"
		devenvConfig.Scripts["build"] = "npm run build"
		devenvConfig.Scripts["start"] = "npm start"
		devenvConfig.Scripts["test"] = "npm test"
		devenvConfig.Scripts["lint"] = "npm run lint"
	}

	// Add TypeScript support
	if config.Options["with_typescript"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages,
			"nodePackages.typescript",
			"nodePackages.ts-node",
			"nodePackages.@types/node",
		)
		devenvConfig.Scripts["tsc"] = "tsc"
		devenvConfig.Scripts["ts-node"] = "ts-node"
		devenvConfig.Scripts["typecheck"] = "tsc --noEmit"
	}

	// Add framework-specific packages and scripts
	framework := config.Options["framework"]
	switch framework {
	case "express":
		devenvConfig.Scripts["create"] = "npx express-generator ."
	case "fastify":
		devenvConfig.Scripts["create"] = "npx fastify-cli generate ."
	case "nextjs":
		devenvConfig.Scripts["create"] = "npx create-next-app@latest ."
		devenvConfig.Scripts["dev"] = "next dev"
		devenvConfig.Scripts["build"] = "next build"
		devenvConfig.Scripts["start"] = "next start"
	case "nuxtjs":
		devenvConfig.Scripts["create"] = "npx nuxi@latest init ."
		devenvConfig.Scripts["dev"] = "nuxt dev"
		devenvConfig.Scripts["build"] = "nuxt build"
		devenvConfig.Scripts["start"] = "nuxt preview"
	case "react":
		devenvConfig.Scripts["create"] = "npx create-react-app ."
	case "vue":
		devenvConfig.Scripts["create"] = "npx create-vue@latest ."
	case "svelte":
		devenvConfig.Scripts["create"] = "npx sv create ."
	}

	// Add common development tools
	devenvConfig.Packages = append(devenvConfig.Packages,
		"nodePackages.eslint",
		"nodePackages.prettier",
		"nodePackages.nodemon",
	)

	devenvConfig.Scripts["format"] = "prettier --write ."
	devenvConfig.Scripts["format-check"] = "prettier --check ."

	// Add requested services
	if len(config.Services) > 0 {
		devenvConfig.Services = make(map[string]interface{})
		for _, service := range config.Services {
			switch service {
			case "postgres":
				devenvConfig.Services["postgres"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "postgres://postgres@localhost/" + config.ProjectName
			case "redis":
				devenvConfig.Services["redis"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Environment["REDIS_URL"] = "redis://localhost:6379"
			case "mysql":
				devenvConfig.Services["mysql"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "mysql://root@localhost/" + config.ProjectName
			case "mongodb":
				devenvConfig.Services["mongodb"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Environment["MONGODB_URL"] = "mongodb://localhost:27017/" + config.ProjectName
			}
		}
	}

	// Add custom packages
	devenvConfig.Packages = append(devenvConfig.Packages, config.Packages...)

	// Add custom environment variables
	for key, value := range config.EnvVars {
		devenvConfig.Environment[key] = value
	}

	return devenvConfig, nil
}

// GolangTemplate implements the Template interface for Go development environments
type GolangTemplate struct{}

func (g *GolangTemplate) Name() string {
	return "golang"
}

func (g *GolangTemplate) Description() string {
	return "Go development environment with go toolchain and common tools"
}

func (g *GolangTemplate) RequiredInputs() []InputField {
	return []InputField{
		{
			Name:        "go_version",
			Type:        "choice",
			Description: "Go version to use",
			Required:    false,
			Default:     "1.21",
			Choices:     []string{"1.19", "1.20", "1.21", "1.22"},
		},
		{
			Name:        "with_air",
			Type:        "bool",
			Description: "Include Air for live reloading",
			Required:    false,
			Default:     "true",
		},
		{
			Name:        "with_grpc",
			Type:        "bool",
			Description: "Include gRPC and Protocol Buffers support",
			Required:    false,
			Default:     "false",
		},
		{
			Name:        "framework",
			Type:        "choice",
			Description: "Web framework (optional)",
			Required:    false,
			Default:     "none",
			Choices:     []string{"none", "gin", "echo", "fiber", "chi"},
		},
	}
}

func (g *GolangTemplate) SupportedServices() []string {
	return []string{"postgres", "redis", "mysql", "mongodb"}
}

func (g *GolangTemplate) Validate(config TemplateConfig) error {
	return nil
}

func (g *GolangTemplate) Generate(config TemplateConfig) (*DevenvConfig, error) {
	goVersion := config.Options["go_version"]
	if goVersion == "" {
		goVersion = "1.21"
	}

	devenvConfig := &DevenvConfig{
		Languages: map[string]interface{}{
			"go": map[string]interface{}{
				"enable":  true,
				"version": goVersion,
			},
		},
		Packages: []string{
			"git",
			"curl",
			"gcc", // Required for CGO
		},
		Environment: map[string]string{
			"CGO_ENABLED": "1",
		},
		Scripts: map[string]interface{}{
			"build":      "go build",
			"run":        "go run .",
			"test":       "go test ./...",
			"fmt":        "go fmt ./...",
			"vet":        "go vet ./...",
			"mod-tidy":   "go mod tidy",
			"mod-vendor": "go mod vendor",
			"clean":      "go clean",
		},
	}

	// Add Air for live reloading
	if config.Options["with_air"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages, "air")
		devenvConfig.Scripts["dev"] = "air"
		devenvConfig.Scripts["air-init"] = "air init"
	}

	// Add gRPC and Protocol Buffers support
	if config.Options["with_grpc"] == "true" {
		devenvConfig.Packages = append(devenvConfig.Packages,
			"protobuf",
			"protoc-gen-go",
			"protoc-gen-go-grpc",
		)
		devenvConfig.Scripts["protoc"] = "protoc"
		devenvConfig.Scripts["proto-gen"] = "protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative"
	}

	// Add framework-specific initialization
	framework := config.Options["framework"]
	switch framework {
	case "gin":
		devenvConfig.Scripts["init-gin"] = "go mod init " + config.ProjectName + " && go get github.com/gin-gonic/gin"
	case "echo":
		devenvConfig.Scripts["init-echo"] = "go mod init " + config.ProjectName + " && go get github.com/labstack/echo/v4"
	case "fiber":
		devenvConfig.Scripts["init-fiber"] = "go mod init " + config.ProjectName + " && go get github.com/gofiber/fiber/v2"
	case "chi":
		devenvConfig.Scripts["init-chi"] = "go mod init " + config.ProjectName + " && go get github.com/go-chi/chi/v5"
	}

	// Add common Go development tools
	devenvConfig.Packages = append(devenvConfig.Packages,
		"golangci-lint", // Linter
		"gopls",         // LSP server
		"delve",         // Debugger
		"go-tools",      // Static analysis tools
	)

	devenvConfig.Scripts["lint"] = "golangci-lint run"
	devenvConfig.Scripts["lint-fix"] = "golangci-lint run --fix"
	devenvConfig.Scripts["debug"] = "dlv debug"

	// Add benchmarking and profiling tools
	devenvConfig.Scripts["bench"] = "go test -bench=."
	devenvConfig.Scripts["prof-cpu"] = "go test -cpuprofile=cpu.prof -bench=."
	devenvConfig.Scripts["prof-mem"] = "go test -memprofile=mem.prof -bench=."

	// Add requested services
	if len(config.Services) > 0 {
		devenvConfig.Services = make(map[string]interface{})
		for _, service := range config.Services {
			switch service {
			case "postgres":
				devenvConfig.Services["postgres"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "postgres://postgres@localhost/" + config.ProjectName
			case "redis":
				devenvConfig.Services["redis"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Environment["REDIS_URL"] = "redis://localhost:6379"
			case "mysql":
				devenvConfig.Services["mysql"] = map[string]interface{}{
					"enable": true,
					"initialDatabases": []map[string]string{
						{"name": config.ProjectName},
					},
				}
				devenvConfig.Environment["DATABASE_URL"] = "mysql://root@localhost/" + config.ProjectName
			case "mongodb":
				devenvConfig.Services["mongodb"] = map[string]interface{}{
					"enable": true,
				}
				devenvConfig.Environment["MONGODB_URL"] = "mongodb://localhost:27017/" + config.ProjectName
			}
		}
	}

	// Add custom packages
	devenvConfig.Packages = append(devenvConfig.Packages, config.Packages...)

	// Add custom environment variables
	for key, value := range config.EnvVars {
		devenvConfig.Environment[key] = value
	}

	return devenvConfig, nil
}
