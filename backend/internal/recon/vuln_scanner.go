package recon

import (
	"fmt"
	"os"
	"path/filepath"

	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunVulnScan(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting vulnerability scan...", "running")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	httpxFile := filepath.Join(resultsDir, "httpx", "httpx-subs.txt")
	if _, err := os.Stat(httpxFile); os.IsNotExist(err) {
		subtask.SetError("httpx-subs.txt not found, please run subdomain probing first")
		logFunc(subtask.Error, "error")
		return
	}

	allSubsFile := filepath.Join(resultsDir, "active", "active-subs.txt")
	if _, err := os.Stat(allSubsFile); os.IsNotExist(err) {
		subtask.SetError("all-subs.txt not found, please run subdomain discovery first")
		logFunc(subtask.Error, "error")
		return
	}

	vulnDir := filepath.Join(resultsDir, "vuln")
	if err := os.MkdirAll(vulnDir, os.ModePerm); err != nil {
		subtask.SetError(fmt.Sprintf("Failed to create directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	subtakeoverFile := filepath.Join(vulnDir, "subtakeover.txt")

	// Run Nuclei with severity-based templates
	severities := []string{"critical", "high", "medium", "low"}
	for _, severity := range severities {
		logFunc(fmt.Sprintf("Running Nuclei scan for %s severity...", severity), "running")
		nucleiOutFile := filepath.Join(vulnDir, fmt.Sprintf("nuclei_%s.txt", severity))
		cmd := fmt.Sprintf("nuclei -l %s -s %s -o %s", httpxFile, severity, nucleiOutFile)
		if err := utils.RunCommand(cmd, logFunc); err != nil {
			logFunc(fmt.Sprintf("Error running Nuclei for %s severity: %v", severity, err), "error")
			// Continue with other severities
		}
	}

	// Run crosy for CORS misconfigurations
	logFunc("Running crosy for CORS misconfigurations...", "running")
	crosyOutFile := filepath.Join(vulnDir, "crosy.txt")
	cmd := fmt.Sprintf("python3 /path/to/crosy.py -i %s -o %s", httpxFile, crosyOutFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running crosy: %v", err), "error")
	}

	// Run Nuclei for social media handles
	logFunc("Running Nuclei for social media handles...", "running")
	socialHandlesFile := filepath.Join(vulnDir, "social-handles.txt")
	cmd = fmt.Sprintf("nuclei -l %s -t social-media -o %s", httpxFile, socialHandlesFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running Nuclei for social media handles: %v", err), "error")
	}

	// Run Nuclei DAST fuzzing
	logFunc("Running Nuclei DAST fuzzing...", "running")
	urlsWithParamsFile := filepath.Join(resultsDir, "urls", "all_urls_with_params.txt")
	if _, err := os.Stat(urlsWithParamsFile); !os.IsNotExist(err) {
		fuzzOutFile := filepath.Join(vulnDir, "nuclei_fuzz.txt")
		cmd = fmt.Sprintf("nuclei -l %s -t fuzz -o %s", urlsWithParamsFile, fuzzOutFile)
		if err := utils.RunCommand(cmd, logFunc); err != nil {
			logFunc(fmt.Sprintf("Error running Nuclei DAST fuzzing: %v", err), "error")
		}
	} else {
		logFunc("No URLs with parameters found for fuzzing.", "")
	}

	// Subdomain takeover scans
	logFunc("Running subjack for subdomain takeover...", "running")
	cmd = fmt.Sprintf("subjack -w %s -t 100 -timeout 30 -ssl -c ~/wordlists/subjack.json -v 3 | grep -v \"Not\" >> %s", allSubsFile, subtakeoverFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running subjack: %v", err), "error")
	}

	logFunc("Running subzy for subdomain takeover...", "running")
	cmd = fmt.Sprintf("subzy run --targets %s --verify_ssl | grep -v \"Unbounce\" | grep VULNERABLE | grep -v NOT >> %s", allSubsFile, subtakeoverFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running subzy: %v", err), "error")
	}

	logFunc("Running Nuclei for subdomain takeover...", "running")
	cmd = fmt.Sprintf("cat %s | nuclei -t ~/nuclei-templates/http/takeovers >> %s", httpxFile, subtakeoverFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running Nuclei for subdomain takeover: %v", err), "error")
	}

	logFunc("Vulnerability scan completed.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
