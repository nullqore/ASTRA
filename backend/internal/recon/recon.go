package recon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
)

// ReconOrchestrator manages the entire reconnaissance process.
type ReconOrchestrator struct {
	ProjectName string
	Modules     []string
	Task        *tasks.Task
	Log         func(string, string)
	ResultsDir  string
	Ctx         context.Context
	Cancel      context.CancelFunc
}

// findProjectRoot walks up from the current directory to find the directory containing go.mod
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Dir(dir), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

// NewReconOrchestrator creates a new orchestrator instance.
func NewReconOrchestrator(projectName string, modules []string, task *tasks.Task, logFunc func(string, string)) (*ReconOrchestrator, error) {
	resultsBaseDir := os.Getenv("SAAM_RESULTS_DIR")
	if resultsBaseDir == "" {
		projectRoot, err := findProjectRoot()
		if err != nil {
			return nil, err
		}
		resultsBaseDir = filepath.Join(projectRoot, "results")
	}

	resultsDir := filepath.Join(resultsBaseDir, projectName)

	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create results directory: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &ReconOrchestrator{
			ProjectName: projectName,
			Modules:     modules,
			Task:        task,
			Log:         logFunc,
			ResultsDir:  resultsDir,
			Ctx:         ctx,
			Cancel:      cancel,
		},
		nil
}

// RunRecon starts the reconnaissance process.
func RunRecon(projectName string, modules []string, task *tasks.Task, logFunc func(string, string)) {
	r, err := NewReconOrchestrator(projectName, modules, task, logFunc)
	if err != nil {
		logFunc(fmt.Sprintf("Error initializing recon: %s", err), "stopped")
		return
	}
	defer r.Cancel() // Ensure context is cancelled when recon finishes

	logFunc(fmt.Sprintf("Starting reconnaissance for project: %s", r.ProjectName), "running")

	for _, module := range r.Modules {
		// Check for stop/pause signals before running each module
		select {
		case <-r.Task.StopChan:
			r.Log("Reconnaissance stopped by user.", "stopped")
			r.Cancel()
			return
		case paused := <-r.Task.PauseChan:
			if paused {
				r.Task.SetStatus(tasks.StatusPaused)
				r.Log("Reconnaissance paused by user.", "paused")
				// Wait for resume signal
				resumed := <-r.Task.PauseChan
				if !resumed { // Check if it was a stop signal instead
					r.Log("Reconnaissance stopped while paused.", "stopped")
					r.Cancel()
					return
				}
				r.Task.SetStatus(tasks.StatusRunning)
				r.Log("Reconnaissance resumed by user.", "running")
			}
		default:
			// Continue execution
		}

		r.Log(fmt.Sprintf("\n--- Running module: %s ---", module), "")
		var modErr error
		switch module {
		case "subfinder":
			modErr = r.runSubdomainDiscovery()
		case "probe":
			modErr = r.runProbe()
		case "port_scan":
			modErr = r.runPortScan()
		case "urls_crawler":
			modErr = r.runURLFinder()
		case "js_crawler":
			modErr = r.runJSScanner()
		case "tech_detect":
			modErr = r.runTechDetection()
		case "paramspyder":
			modErr = r.runHiddenParameter()
		case "fuzzer":
			modErr = r.runFuzzer()
		case "vuln_scan":
			modErr = r.runVulnScan()
		case "xss_scan":
			modErr = r.runXSSScan()
		case "sqli_scan":
			modErr = r.runSQLiScan()
		case "screenshot":
			modErr = r.runScreenshot()
		default:
			r.Log(fmt.Sprintf("Module '%s' is not yet implemented.", module), "")
		}
		if modErr != nil {
			if modErr == context.Canceled {
				r.Log(fmt.Sprintf("Module %s stopped by user.", module), "stopped")
				return
			}
			r.Log(fmt.Sprintf("Error running module %s: %s", module, modErr), "")
		}
	}

	r.Log("\n--- Reconnaissance complete ---", "stopped")
}

func (r *ReconOrchestrator) runURLFinder() error {
	taskKey := "urls"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "URL Finder", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting URL finder...", "running")

	RunURLFinder(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("URL finder completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runJSScanner() error {
	taskKey := "js"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "JS Scanner", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting JS scanner...", "running")

	RunJSScanner(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("JS scanner completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runTechDetection() error {
	taskKey := "tech_detect"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Tech Detection", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting tech detection...", "running")

	RunTechDetection(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Tech detection completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runHiddenParameter() error {
	taskKey := "hidden_parameter"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Hidden Parameter", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting hidden parameter scan...", "running")

	RunParamSpider(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Hidden parameter scan completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runFuzzer() error {
	taskKey := "fuzzer"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Fuzzer", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting fuzzer...", "running")

	RunFuzzer(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Fuzzer completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runVulnScan() error {
	taskKey := "vuln_scan"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Vulnerability Scan", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting vulnerability scan...", "running")

	RunVulnScan(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Vulnerability scan completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runXSSScan() error {
	taskKey := "xss_scan"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "XSS Scan", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting XSS scan...", "running")

	RunXSSScan(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("XSS scan completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runScreenshot() error {
	taskKey := "screenshot"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Screenshot", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting screenshotting...", "running")

	RunScreenshot(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Screenshotting completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runSQLiScan() error {
	taskKey := "sqli_scan"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "SQLi Scan", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting SQLi scan...", "running")

	RunSQLiScan(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("SQLi scan completed.", "completed")
	return nil
}

func (r *ReconOrchestrator) runPortScan() error {
	taskKey := "port_scan"
	subtask := r.Task.GetSubtask(taskKey)
	if subtask == nil {
		subtask = &tasks.Subtask{Name: "Port Scan", Status: tasks.StatusPending}
		r.Task.AddSubtask(taskKey, subtask)
	}
	r.Log("Starting port scan...", "running")

	RunPortScan(r.ProjectName, subtask, r.Log)

	subtask.SetStatus(tasks.StatusCompleted)
	r.Log("Port scan completed.", "completed")
	return nil
}
