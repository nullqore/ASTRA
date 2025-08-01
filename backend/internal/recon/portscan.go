package recon

import (
	"fmt"
	"os"
	"path/filepath"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunPortScan(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	logFunc("Starting port scan...", "running")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	activeSubsFile := filepath.Join(resultsDir, "active", "active-subs.txt")
	if _, err := os.Stat(activeSubsFile); os.IsNotExist(err) {
		subtask.SetError("active-subs.txt not found, please run subdomain discovery first")
		logFunc(subtask.Error, "error")
		return
	}

	infoDir := filepath.Join(resultsDir, "info")
	portDir := filepath.Join(resultsDir, "port")
	tempDir := filepath.Join(portDir, "temp")
	vulnDir := filepath.Join(resultsDir, "vuln")
	fuffDir := filepath.Join(resultsDir, "fuff")

	for _, dir := range []string{infoDir, portDir, tempDir, vulnDir, fuffDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			subtask.SetError(fmt.Sprintf("Failed to create directory: %v", err))
			logFunc(subtask.Error, "error")
			return
		}
	}

	cnameFile := filepath.Join(infoDir, "cname-subs.txt")
	tempAllIPsFile := filepath.Join(tempDir, "temp_all-ips.txt")
	noCDNIPsFile := filepath.Join(portDir, "nocdn-ips.txt")
	finalIPsFile := filepath.Join(portDir, "final-ips.txt")
	nmapOutFile := filepath.Join(portDir, "nmap_out.txt")
	naabuOutFile := filepath.Join(portDir, "naabu_out.txt")
	massdnsOutFile := filepath.Join(portDir, "massdns_out.txt")
	nucleiOutFile := filepath.Join(vulnDir, "ips_nuclei_out.txt")

	logFunc("Finding CNAMEs...", "running")
	cmd := fmt.Sprintf("cat %s | dnsx -silent -cname -resp | anew -q %s", activeSubsFile, cnameFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error finding CNAMEs: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Finding origin IPs...", "running")
	cmd = fmt.Sprintf(`cat %s | sed 's/^-//' | xargs -P 50 -I {} dig +short {} A | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" | sort -u | anew -q %s`, activeSubsFile, tempAllIPsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error finding origin IPs with dig: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	cmd = fmt.Sprintf("dnsx -l %s -resp-only | anew -q %s", activeSubsFile, tempAllIPsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error finding origin IPs with dnsx: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Filtering out CDN IPs...", "running")
	cmd = fmt.Sprintf("cat %s | cut-cdn -o %s", tempAllIPsFile, noCDNIPsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error filtering CDN IPs: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Finding virtual hosts...", "running")
	cmd = fmt.Sprintf(`for ip in $(cat %s);do echo $ip &&  ffuf -w %s -u http://$ip -H "Host: FUZZ" -s -mc 200; done | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" |  tee %s`, noCDNIPsFile, activeSubsFile, finalIPsFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error finding virtual hosts: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Running nmap scan...", "running")
	cmd = fmt.Sprintf("nmap -iL %s -oN %s", finalIPsFile, nmapOutFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error running nmap: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Running naabu scan...", "running")
	cmd = fmt.Sprintf("naabu -iL %s -o %s", finalIPsFile, naabuOutFile)
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error running naabu: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Running massdns scan...", "running")
	cmd = fmt.Sprintf("massdns -r /path/to/resolvers.txt -t A -o S %s > %s", finalIPsFile, massdnsOutFile) // Assumes resolvers.txt path
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error running massdns: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Running nuclei scan on IPs...", "running")
	cmd = fmt.Sprintf("nuclei -l %s -t /path/to/nuclei-templates -severity low,medium,high,critical -o %s", finalIPsFile, nucleiOutFile) // Assumes nuclei-templates path
	if err := utils.RunCommand(cmd, logFunc); err != nil {
		subtask.SetError(fmt.Sprintf("Error running nuclei: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	logFunc("Running ffuf fuzzing...", "running")
	commonWordlist := "/path/to/common.txt" // Assumes common.txt path
	ips, err := utils.ReadFileLines(finalIPsFile)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error reading final IPs: %v", err))
		logFunc(subtask.Error, "error")
		return
	}

	for _, ip := range ips {
		ffufOutFile := filepath.Join(fuffDir, fmt.Sprintf("%s.txt", ip))
		cmd = fmt.Sprintf("ffuf -w %s -u http://%s/FUZZ -o %s", commonWordlist, ip, ffufOutFile)
		if err := utils.RunCommand(cmd, logFunc); err != nil {
			logFunc(fmt.Sprintf("Error running ffuf on %s: %v", ip, err), "error")
			// Continue with other IPs
		}
	}

	logFunc("Port scan completed.", "completed")
	subtask.SetStatus(tasks.StatusCompleted)
}
