package recon

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

// DeduplicateAndSort removes duplicate strings from a slice and sorts it.
func DeduplicateAndSort(input []string) []string {
	uniqueMap := make(map[string]struct{})
	for _, item := range input {
		if strings.TrimSpace(item) != "" {
			uniqueMap[item] = struct{}{}
		}
	}
	result := make([]string, 0, len(uniqueMap))
	for item := range uniqueMap {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

// RunJSScanner orchestrates a comprehensive JavaScript file analysis pipeline.
func RunJSScanner(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	defer subtask.SetStatus(tasks.StatusCompleted)

	logFunc("🔥 Starting comprehensive JS analysis pipeline...", "")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(fmt.Sprintf("⚠️ Error getting results directory: %v", err), "")
		return
	}

	// Define file paths
	httpxFile := filepath.Join(resultsDir, "httpx", "httpx-subs.txt")
	allUrlsFile := filepath.Join(resultsDir, "urls", "all_urls.txt")
	jsUrlsFile := filepath.Join(resultsDir, "urls", "js-urls.txt")
	activeJsUrlsFile := filepath.Join(resultsDir, "urls", "active-js-urls.txt")
	infoDir := filepath.Join(resultsDir, "info")
	vulnDir := filepath.Join(resultsDir, "vuln")
	endpointFile := filepath.Join(infoDir, "endpoint.txt")
	goendpointFile := filepath.Join(infoDir, "goendpoint.txt")
	mantraOutFile := filepath.Join(vulnDir, "mantra-out.txt")
	nucleiOutFile := filepath.Join(vulnDir, "nuclei-exposure-out.txt")

	// Create necessary directories
	os.MkdirAll(filepath.Join(resultsDir, "urls"), os.ModePerm)
	os.MkdirAll(infoDir, os.ModePerm)
	os.MkdirAll(vulnDir, os.ModePerm)

	// --- Step 1: Gather all potential JS URLs ---
	logFunc("🔍 Step 1/4: Gathering all potential JS URLs...", "running")
	var allJsUrls []string

	// Tool 1: getjs
	if _, err := os.Stat(httpxFile); err == nil {
		logFunc("  -> Running getjs on httpx-subs.txt...", "")
		cmd := exec.Command("bash", "-c", fmt.Sprintf("getJS -input %s --complete", httpxFile))
		output, err := cmd.CombinedOutput()
		if err != nil {
			logFunc(fmt.Sprintf("  -> ⚠️ Error running getjs: %v. Output: %s", err, string(output)), "")
		} else {
			urls := strings.Split(string(output), "\n")
			allJsUrls = append(allJsUrls, urls...)
			logFunc(fmt.Sprintf("  -> ✅ getjs found %d URLs.", len(urls)), "")
		}
	} else {
		logFunc("  -> ℹ️ httpx-subs.txt not found, skipping getjs.", "")
	}

	// Tool 2: subjs
	if _, err := os.Stat(httpxFile); err == nil {
		logFunc("  -> Running subjs on httpx-subs.txt...", "")
		cmd := exec.Command("bash", "-c", fmt.Sprintf("subjs -i %s", httpxFile))
		output, err := cmd.CombinedOutput()
		if err != nil {
			logFunc(fmt.Sprintf("  -> ⚠️ Error running subjs: %v. Output: %s", err, string(output)), "")
		} else {
			urls := strings.Split(string(output), "\n")
			allJsUrls = append(allJsUrls, urls...)
			logFunc(fmt.Sprintf("  -> ✅ subjs found %d URLs.", len(urls)), "")
		}
	} else {
		logFunc("  -> ℹ️ httpx-subs.txt not found, skipping subjs.", "")
	}

	// Tool 3: Grep from all_urls.txt
	if _, err := os.Stat(allUrlsFile); err == nil {
		logFunc("  -> Grepping for .js files in all_urls.txt...", "")
		cmd := exec.Command("bash", "-c", fmt.Sprintf(`cat %s | grep -P '\w+\.js(\?|$)'`, allUrlsFile))
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Don't log error if grep just doesn't find anything
		} else {
			urls := strings.Split(string(output), "\n")
			allJsUrls = append(allJsUrls, urls...)
			logFunc(fmt.Sprintf("  -> ✅ Grep found %d JS URLs.", len(urls)), "")
		}
	} else {
		logFunc("  -> ℹ️ all_urls.txt not found, skipping grep.", "")
	}

	// Deduplicate and write to file
	logFunc(fmt.Sprintf("  -> ℹ️ Found %d total JS URLs. Deduplicating...", len(allJsUrls)), "")
	uniqueJsUrls := DeduplicateAndSort(allJsUrls)
	logFunc(fmt.Sprintf("  -> ℹ️ Found %d unique JS URLs.", len(uniqueJsUrls)), "")
	utils.WriteFileLines(jsUrlsFile, uniqueJsUrls)
	logFunc("✅ Step 1/4: JS URL gathering complete.", "completed")

	// --- Step 2: Probe for active JS URLs ---
	logFunc("🚀 Step 2/4: Probing for active JS URLs...", "running")
	activeJsUrls := probeJSURLs(uniqueJsUrls, logFunc)
	utils.WriteFileLines(activeJsUrlsFile, activeJsUrls)
	logFunc(fmt.Sprintf("✅ Step 2/4: Probing complete. Found %d active JS URLs.", len(activeJsUrls)), "completed")

	// --- Step 3: Find Endpoints & Info ---
	logFunc("🕵️ Step 3/4: Finding endpoints and info from active JS files...", "running")

	// Tool 4: linkfinder
	logFunc("  -> Running linkfinder on active JS URLs...", "")
	linkfinderCmd := fmt.Sprintf("linkfinder -i %s -o cli | anew %s", activeJsUrlsFile, endpointFile)
	cmd := exec.Command("bash", "-c", linkfinderCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logFunc(fmt.Sprintf("  -> ⚠️ Error running linkfinder: %v. Output: %s", err, string(output)), "")
	} else {
		logFunc("  -> ✅ linkfinder completed.", "")
	}

	// Tool 5: Custom js-scanner.py
	logFunc("  -> Running golinkfinder on active JS URLs...", "")
	jsScannerCmd := fmt.Sprintf("golinkfinder -l %s -o %s", activeJsUrlsFile, goendpointFile)
	cmd = exec.Command("bash", "-c", jsScannerCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logFunc(fmt.Sprintf("  -> ⚠️ Error running golinkfinder: %v. Output: %s", err, string(output)), "")
	} else {
		logFunc("  -> ✅ golinkfinder completed.", "")
	}
	logFunc("✅ Step 3/4: Endpoint and info gathering complete.", "completed")

	// --- Step 4: Vulnerability Scanning ---
	logFunc("💥 Step 4/4: Scanning for vulnerabilities...", "running")

	// Tool 6: mantra
	logFunc("  -> Running mantra on active JS URLs...", "")
	mantraCmd := fmt.Sprintf("cat %s | mantra > %s", activeJsUrlsFile, mantraOutFile)
	cmd = exec.Command("bash", "-c", mantraCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logFunc(fmt.Sprintf("  -> ⚠️ Error running mantra: %v. Output: %s", err, string(output)), "")
	} else {
		logFunc(fmt.Sprintf("  -> ✅ mantra scan complete. Results saved to %s", mantraOutFile), "")
	}

	// Tool 7: nuclei
	logFunc("  -> Running nuclei with exposure templates...", "")
	nucleiCmd := fmt.Sprintf("nuclei -l %s -t ~/nuclei-templates/http/exposures -o %s", activeJsUrlsFile, nucleiOutFile)
	cmd = exec.Command("bash", "-c", nucleiCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logFunc(fmt.Sprintf("  -> ⚠️ Error running nuclei: %v. Output: %s", err, string(output)), "")
	} else {
		logFunc(fmt.Sprintf("  -> ✅ nuclei scan complete. Results saved to %s", nucleiOutFile), "")
	}
	logFunc("✅ Step 4/4: Vulnerability scanning complete.", "completed")

	logFunc("🎉 All JS analysis tasks completed successfully!", "completed")
}

// probeJSURLs checks a list of URLs and returns the ones that respond with a 200 OK status.
func probeJSURLs(urls []string, logFunc func(string, string)) []string {
	var activeURLs []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	var probedCount int32
	totalURLs := len(urls)

	transport := &http.Transport{
		MaxIdleConns:        50,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	concurrencyLimit := 50
	guard := make(chan struct{}, concurrencyLimit)

	for _, url := range urls {
		if strings.TrimSpace(url) == "" {
			atomic.AddInt32(&probedCount, 1)
			continue
		}

		wg.Add(1)
		guard <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-guard }()

			resp, err := client.Get(u)
			if err == nil && resp.StatusCode == http.StatusOK {
				mu.Lock()
				activeURLs = append(activeURLs, u)
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}

			newCount := atomic.AddInt32(&probedCount, 1)
			logFunc(fmt.Sprintf("\r  -> Probed %d/%d URLs...", newCount, totalURLs), "")
		}(url)
	}

	wg.Wait()
	logFunc("\n", "")
	sort.Strings(activeURLs)
	return activeURLs
}
