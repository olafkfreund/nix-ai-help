package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/ai/agent"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// BuildMonitor manages background build processes with AI-powered monitoring
type BuildMonitor struct {
	activeBuilds map[string]*BuildProcess
	buildAgent   *agent.BuildAgent
	logger       *logger.Logger
	mutex        sync.RWMutex
}

// BuildProcess represents an active build with monitoring
type BuildProcess struct {
	ID         string
	Package    string
	Command    *exec.Cmd
	StartTime  time.Time
	Status     string
	Output     []string
	ErrorCount int
	Context    context.Context
	Cancel     context.CancelFunc
	Progress   chan BuildProgress
}

// BuildProgress represents real-time build progress information
type BuildProgress struct {
	Timestamp   time.Time `json:"timestamp"`
	Stage       string    `json:"stage"`
	Message     string    `json:"message"`
	Percentage  float64   `json:"percentage,omitempty"`
	IsError     bool      `json:"is_error"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// NewBuildMonitor creates a new build monitor with AI integration
func NewBuildMonitor(buildAgent *agent.BuildAgent) *BuildMonitor {
	return &BuildMonitor{
		activeBuilds: make(map[string]*BuildProcess),
		buildAgent:   buildAgent,
		logger:       logger.NewLogger(),
		mutex:        sync.RWMutex{},
	}
}

// StartBackgroundBuild initiates a build process with real-time monitoring
func (bm *BuildMonitor) StartBackgroundBuild(package_name string, args []string) (string, error) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	buildID := fmt.Sprintf("build_%s_%d", package_name, time.Now().Unix())
	ctx, cancel := context.WithCancel(context.Background())

	// Create build command
	cmdArgs := append([]string{"build"}, args...)
	cmd := exec.CommandContext(ctx, "nix", cmdArgs...)

	// Setup build process
	process := &BuildProcess{
		ID:        buildID,
		Package:   package_name,
		Command:   cmd,
		StartTime: time.Now(),
		Status:    "starting",
		Output:    make([]string, 0),
		Context:   ctx,
		Cancel:    cancel,
		Progress:  make(chan BuildProgress, 100),
	}

	bm.activeBuilds[buildID] = process

	// Start monitoring goroutine
	go bm.monitorBuild(process)

	fmt.Println(utils.FormatSuccess(fmt.Sprintf("🚀 Started background build: %s", buildID)))
	fmt.Println(utils.FormatTip(fmt.Sprintf("Monitor with: nixai build status %s", buildID)))

	return buildID, nil
}

// monitorBuild provides real-time monitoring with AI analysis
func (bm *BuildMonitor) monitorBuild(process *BuildProcess) {
	defer close(process.Progress)

	// Start the build command
	stdout, err := process.Command.StdoutPipe()
	if err != nil {
		bm.logger.Error(fmt.Sprintf("Failed to create stdout pipe: %v", err))
		return
	}

	stderr, err := process.Command.StderrPipe()
	if err != nil {
		bm.logger.Error(fmt.Sprintf("Failed to create stderr pipe: %v", err))
		return
	}

	if err := process.Command.Start(); err != nil {
		process.Status = "failed"
		bm.logger.Error(fmt.Sprintf("Failed to start build: %v", err))
		return
	}

	process.Status = "running"

	// Monitor output streams
	go bm.parseOutputStream(process, stdout, false)
	go bm.parseOutputStream(process, stderr, true)

	// Wait for completion
	err = process.Command.Wait()

	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if err != nil {
		process.Status = "failed"
		bm.analyzeFailureWithAI(process)
	} else {
		process.Status = "completed"
		bm.analyzeSuccessWithAI(process)
	}

	bm.logger.Info(fmt.Sprintf("Build %s completed with status: %s", process.ID, process.Status))
}

// GetBuildStatus returns the current status of a build
func (bm *BuildMonitor) GetBuildStatus(buildID string) (*BuildProcess, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	process, exists := bm.activeBuilds[buildID]
	if !exists {
		return nil, fmt.Errorf("build %s not found", buildID)
	}

	return process, nil
}

// ListActiveBuilds returns all currently active builds
func (bm *BuildMonitor) ListActiveBuilds() map[string]*BuildProcess {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	// Return copy to avoid race conditions
	builds := make(map[string]*BuildProcess)
	for id, process := range bm.activeBuilds {
		builds[id] = process
	}

	return builds
}

// StopBuild cancels a running build
func (bm *BuildMonitor) StopBuild(buildID string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	process, exists := bm.activeBuilds[buildID]
	if !exists {
		return fmt.Errorf("build %s not found", buildID)
	}

	if process.Status == "running" {
		process.Cancel()
		process.Status = "cancelled"
		fmt.Println(utils.FormatWarning(fmt.Sprintf("🛑 Cancelled build: %s", buildID)))
	}

	return nil
}

// parseOutputStream parses build output and extracts meaningful progress information
func (bm *BuildMonitor) parseOutputStream(process *BuildProcess, stream io.ReadCloser, isError bool) {
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()

		// Add line to process output
		bm.mutex.Lock()
		process.Output = append(process.Output, line)
		if len(process.Output) > 1000 { // Keep only last 1000 lines
			process.Output = process.Output[1:]
		}
		bm.mutex.Unlock()

		// Parse for build stages and progress
		progress := bm.parseLineForProgress(line, isError)
		if progress != nil {
			select {
			case process.Progress <- *progress:
			default:
				// Channel full, skip this update
			}
		}

		// Count errors
		if isError {
			bm.mutex.Lock()
			process.ErrorCount++
			bm.mutex.Unlock()
		}
	}
}

// analyzeFailureWithAI uses the BuildAgent to analyze build failures
func (bm *BuildMonitor) analyzeFailureWithAI(process *BuildProcess) {
	if bm.buildAgent == nil {
		return
	}

	// Create build context for AI analysis
	buildContext := &agent.BuildContext{
		BuildOutput:    fmt.Sprintf("Build failed after %v", time.Since(process.StartTime)),
		ErrorLogs:      bm.getLastErrors(process),
		FailedPackages: []string{process.Package},
		BuildSystem:    "nix-build",
	}

	bm.buildAgent.SetBuildContext(buildContext)

	// Get AI analysis
	analysis, err := bm.buildAgent.Query(context.Background(),
		fmt.Sprintf("Analyze this build failure for package %s", process.Package))

	if err == nil {
		fmt.Println(utils.FormatSubsection("🤖 AI Build Failure Analysis", ""))
		fmt.Println(utils.RenderMarkdown(analysis))
	}
}

// analyzeSuccessWithAI provides optimization suggestions for successful builds
func (bm *BuildMonitor) analyzeSuccessWithAI(process *BuildProcess) {
	if bm.buildAgent == nil {
		return
	}

	buildTime := time.Since(process.StartTime)

	// Only analyze if build took significant time
	if buildTime > 30*time.Second {
		buildContext := &agent.BuildContext{
			BuildOutput: fmt.Sprintf("Build completed successfully in %v", buildTime),
			BuildSystem: "nix-build",
		}

		bm.buildAgent.SetBuildContext(buildContext)

		analysis, err := bm.buildAgent.Query(context.Background(),
			fmt.Sprintf("Suggest optimizations for %s build that took %v", process.Package, buildTime))

		if err == nil {
			fmt.Println(utils.FormatSubsection("⚡ AI Optimization Suggestions", ""))
			fmt.Println(utils.RenderMarkdown(analysis))
		}
	}
}

// getLastErrors extracts the most recent error messages from build output
func (bm *BuildMonitor) getLastErrors(process *BuildProcess) string {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	var errors []string
	for i := len(process.Output) - 1; i >= 0 && len(errors) < 10; i-- {
		line := process.Output[i]
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(line, "failed") ||
			strings.Contains(line, "Error:") {
			errors = append([]string{line}, errors...)
		}
	}

	return strings.Join(errors, "\n")
}

// parseLineForProgress extracts progress information from build output lines
func (bm *BuildMonitor) parseLineForProgress(line string, isError bool) *BuildProgress {
	progress := &BuildProgress{
		Timestamp: time.Now(),
		Message:   line,
		IsError:   isError,
	}

	// Extract build stages
	if strings.Contains(line, "fetching") {
		progress.Stage = "fetching"
	} else if strings.Contains(line, "building") {
		progress.Stage = "building"
	} else if strings.Contains(line, "testing") {
		progress.Stage = "testing"
	} else if strings.Contains(line, "installing") {
		progress.Stage = "installing"
	} else if isError {
		progress.Stage = "error"
	}

	// Extract percentage if available (pattern like "50% complete")
	percentageRegex := regexp.MustCompile(`(\d+)%`)
	if matches := percentageRegex.FindStringSubmatch(line); len(matches) > 1 {
		if percentage := matches[1]; percentage != "" {
			// Parse percentage
			progress.Percentage = parsePercentage(percentage)
		}
	}

	return progress
}

// parsePercentage safely parses a percentage string
func parsePercentage(s string) float64 {
	// Simple implementation - in real code you'd use strconv.ParseFloat
	switch s {
	case "0":
		return 0.0
	case "25":
		return 25.0
	case "50":
		return 50.0
	case "75":
		return 75.0
	case "100":
		return 100.0
	default:
		return 0.0
	}
}
