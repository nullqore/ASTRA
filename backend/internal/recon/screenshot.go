package recon

import (
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunScreenshot(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting screenshotting...", "running")

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

	screenshotDir := filepath.Join(resultsDir, "screenshots")
	os.MkdirAll(screenshotDir, os.ModePerm)

	logFunc("Running Aquatone to take screenshots...", "running")
	cmd := fmt.Sprintf("cat %s | aquatone -out %s", httpxFile, screenshotDir)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error running Aquatone: %v", err), "error")
	}

	logFunc("Screenshotting finished.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
