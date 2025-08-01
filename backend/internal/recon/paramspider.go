package recon

import (
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunParamSpider(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting parameter discovery...", "running")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	urlsDir := filepath.Join(resultsDir, "urls")
	finalUrlsFile := filepath.Join(urlsDir, "final-urls.txt")

	if _, err := os.Stat(finalUrlsFile); os.IsNotExist(err) {
		logFunc("final-urls.txt not found, please run URL discovery first.", "error")
		subtask.SetError("final-urls.txt not found")
		return
	}

	infoDir := filepath.Join(resultsDir, "info")
	if err := os.MkdirAll(infoDir, os.ModePerm); err != nil {
		subtask.SetError(fmt.Sprintf("Failed to create directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	paramlistFile := filepath.Join(infoDir, "paramlist.txt")
	phpUrlsFile := filepath.Join(urlsDir, "php-urls.txt")
	aspxUrlsFile := filepath.Join(urlsDir, "aspx-urls.txt")
	jspUrlsFile := filepath.Join(urlsDir, "jsp-urls.txt")
	ashxUrlsFile := filepath.Join(urlsDir, "ashx-urls.txt")
	cgiUrlsFile := filepath.Join(urlsDir, "cgi-urls.txt")
	xmlUrlsFile := filepath.Join(urlsDir, "xml-urls.txt")
	txtUrlsFile := filepath.Join(urlsDir, "txt-urls.txt")
	xhtmlUrlsFile := filepath.Join(urlsDir, "xhtml-urls.txt")

	logFunc("Extracting params from URLs...", "running")
	cmd := fmt.Sprintf("cat %s | sort -u | unfurl --unique keys | anew -q %s", finalUrlsFile, paramlistFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error extracting params: %v", err), "error")
	}

	logFunc("Filtering PHP URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.php(\?|$)" | sort -u | anew -q %s`, finalUrlsFile, phpUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering PHP URLs: %v", err), "error")
	}

	logFunc("Filtering ASPX URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.aspx(\?|$)" | sort -u | anew -q %s`, finalUrlsFile, aspxUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering ASPX URLs: %v", err), "error")
	}

	logFunc("Filtering JSP URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.jsp(\?|$)" | sort -u | anew -q %s`, finalUrlsFile, jspUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering JSP URLs: %v", err), "error")
	}

	logFunc("Filtering ASHX URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.ashx(\?|$)" | sort -u | anew -q %s`, finalUrlsFile, ashxUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering ASHX URLs: %v", err), "error")
	}

	logFunc("Filtering CGI URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.cgi\?|$)" | sort -u | anew -q %s`, finalUrlsFile, cgiUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering CGI URLs: %v", err), "error")
	}

	logFunc("Filtering XML URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.xml(\?|$)" | sort -u | anew -q %s`, finalUrlsFile, xmlUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering XML URLs: %v", err), "error")
	}

	logFunc("Filtering TXT URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.txt\?|$)" | sort -u | anew -q %s`, finalUrlsFile, txtUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering TXT URLs: %v", err), "error")
	}

	logFunc("Filtering XHTML URLs...", "running")
	cmd = fmt.Sprintf(`cat %s | grep -P "\w+\.xhtml?|$)" | sort -u | anew -q %s`, finalUrlsFile, xhtmlUrlsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		logFunc(fmt.Sprintf("Error filtering XHTML URLs: %v", err), "error")
	}

	logFunc("Parameter discovery finished.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
