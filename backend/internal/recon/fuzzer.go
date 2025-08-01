package recon

import (
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunFuzzer(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting fuzzer...", "running")

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

	fuzzerDir := filepath.Join(resultsDir, "fuzzer")
	os.MkdirAll(fuzzerDir, os.ModePerm)

	// Fuzz for common files and directories
	logFunc("Fuzzing for common files and directories...", "running")
	commonWordlist := "/path/to/common.txt" // Assumes common.txt path
	ffufCommonOutFile := filepath.Join(fuzzerDir, "ffuf_common.txt")
	cmd := fmt.Sprintf("ffuf -w %s -u %s/FUZZ -o %s", commonWordlist, httpxFile, ffufCommonOutFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running ffuf with common.txt: %v", err), "error")
	}

	logFunc("Fuzzer finished.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
