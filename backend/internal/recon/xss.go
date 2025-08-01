package recon

import (
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunXSSScan(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting XSS scan...", "running")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	// Define directories
	urlsDir := filepath.Join(resultsDir, "urls")
	tempUrlsDir := filepath.Join(urlsDir, "temp")
	xssDir := filepath.Join(resultsDir, "vuln", "xss")
	os.MkdirAll(xssDir, os.ModePerm)

	// Combine all URL files into one
	allUrlsFile := filepath.Join(urlsDir, "parameter.txt")
	logFunc("Combining all found URLs with parameters...", "running")
	if _, err := os.Stat(tempUrlsDir); os.IsNotExist(err) {
		subtask.SetError("temp URL directory not found, please run URL discovery first")
		logFunc(subtask.Error, "error")
		return
	}

	combinedCmd := fmt.Sprintf(`find %s -type f -name '*.txt' -exec cat {} + | grep '=' | anew %s`, tempUrlsDir, allUrlsFile)
	if err := utils.RunCommand(combinedCmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error combining URL files: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	fileInfo, err := os.Stat(allUrlsFile)
	if err != nil || fileInfo.Size() == 0 {
		logFunc("No URLs with parameters found to scan for XSS.", "completed")
		subtask.SetStatus(tasks.StatusCompleted)
		return
	}

	// 1. One-liner XSS checks
	logFunc("Running initial XSS one-liners...", "running")
	onelinerOutFile := filepath.Join(xssDir, "oneliner_xss_output.txt")
	cmd1 := fmt.Sprintf(`cat %s | qsreplace '"/><script>alert(1)</script>' | freq | tee %s`, allUrlsFile, onelinerOutFile)
	if err := utils.RunCommand(cmd1, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running XSS one-liner: %v", err), "error")
	}

	// 2. Run Dalfox
	logFunc("Running Dalfox for deeper XSS scanning...", "running")
	dalfoxOutFile := filepath.Join(xssDir, "dalfox_output.txt")
	cmd2 := fmt.Sprintf(`dalfox file %s -o %s`, allUrlsFile, dalfoxOutFile)
	if err := utils.RunCommand(cmd2, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running Dalfox: %v", err), "error")
	}

	// 3. Run Gxss to find reflections
	logFunc("Running Gxss to find reflection points...", "running")
	reflectionUrlsFile := filepath.Join(urlsDir, "reflection_urls.txt")
	cmd3 := fmt.Sprintf(`cat %s | Gxss -p '"/><script>alert(1)</script>' -o %s`, allUrlsFile, reflectionUrlsFile)
	if err := utils.RunCommand(cmd3, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running Gxss: %v", err), "error")
	}

	// 4. Run Dalfox on reflected URLs
	if _, err := os.Stat(reflectionUrlsFile); err == nil {
		logFunc("Running Dalfox on reflected URLs...", "running")
		dalfoxReflectedOutFile := filepath.Join(xssDir, "dalfox_reflected_output.txt")
		cmd4 := fmt.Sprintf(`dalfox file %s -o %s`, reflectionUrlsFile, dalfoxReflectedOutFile)
		if err := utils.RunCommand(cmd4, logFunc); err != nil {
			logFunc(fmt.Sprintf("Error running Dalfox on reflected URLs: %v", err), "error")
		}
	}

	// 5. Run XSStrike
	logFunc("Running XSStrike for payload fuzzing...", "running")
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting project root: %v", err))
		logFunc(subtask.Error, "error")
		return
	}
	payloadFile := filepath.Join(projectRoot, "data", "xss-wordlist.txt")
	xsstrikeOutFile := filepath.Join(xssDir, "xsstrike_output.txt")
	cmd5 := fmt.Sprintf(`python3 /path/to/xsstrike.py -u "$(cat %s)" --fuzzer %s > %s`, allUrlsFile, payloadFile, xsstrikeOutFile)
	if err := utils.RunCommand(cmd5, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running XSStrike: %v", err), "error")
	}

	logFunc("XSS scan completed.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
