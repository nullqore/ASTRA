package recon

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"project-saam/backend/internal/tasks"
	"project-saam/backend/pkg/utils"
)

func RunURLFinder(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	defer subtask.SetStatus(tasks.StatusCompleted)

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		return
	}

	urlsDir := filepath.Join(resultsDir, "urls")
	os.MkdirAll(urlsDir, os.ModePerm)

	tempUrlsDir := filepath.Join(urlsDir, "temp")
	os.MkdirAll(tempUrlsDir, os.ModePerm)

	activeFile := filepath.Join(resultsDir, "active", "active-subs.txt")
	httpxFile := filepath.Join(resultsDir, "httpx", "httpx-subs.txt")

	activeTools := []string{"gau", "waybackurls", "urlfinder", "github-endpoints", "cariddi", "gourlex", "orwa", "waymore"}
	httpxTools := []string{"hakrawler", "gospider"}

	// Process active.txt sequentially
	if _, err := os.Stat(activeFile); os.IsNotExist(err) {
		logFunc(fmt.Sprintf("ℹ️ active.txt not found, skipping associated URL tools."), "")
	} else {
		logFunc("🔥 Starting URL discovery on active-subs.txt...", "")
		err := runToolsOnFile(projectName, activeFile, activeTools, logFunc)
		if err != nil {
			logFunc(fmt.Sprintf("⚠️ Error processing active.txt: %v", err), "")
		} else {
			logFunc("✅ Finished URL discovery on active.txt.", "")
		}
	}

	// Process httpx-subs.txt sequentially
	if _, err := os.Stat(httpxFile); os.IsNotExist(err) {
		logFunc(fmt.Sprintf("ℹ️ httpx-subs.txt not found, skipping associated URL tools."), "")
	} else {
		logFunc("🔥 Starting URL discovery on httpx-subs.txt...", "")
		err := runToolsOnFile(projectName, httpxFile, httpxTools, logFunc)
		if err != nil {
			logFunc(fmt.Sprintf("⚠️ Error processing httpx-subs.txt: %v", err), "")
		} else {
			logFunc("✅ Finished URL discovery on httpx-subs.txt.", "")
		}
	}

	// Aggregate results from files using a shell command
	logFunc("Aggregating and sorting results from files...", "")
	urlsFile := filepath.Join(urlsDir, "all_urls.txt")
	cmdStr := fmt.Sprintf("cat %s/*_urls.txt | strings | grep -E 'http|https' | sort -u > %s", tempUrlsDir, urlsFile)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error aggregating URLs: %v. Output: %s", err, string(output)))
		return
	}

	// Read the aggregated URLs for probing
	content, err := os.ReadFile(urlsFile)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error reading aggregated URLs file: %v", err))
		return
	}
	uniqueURLs := strings.Split(string(content), "\n")
	logFunc(fmt.Sprintf("ℹ️ Found %d unique URLs.", len(uniqueURLs)), "")

	// Probe unique URLs
	logFunc(fmt.Sprintf("🚀 Probing %d unique URLs with maximum accuracy...", len(uniqueURLs)), "")
	activeURLs := probeURLs(uniqueURLs, logFunc)
	logFunc(fmt.Sprintf("✅ Found %d active URLs after final probing.", len(activeURLs)), "")

	// Write active URLs to file
	activeUrlsFile := filepath.Join(urlsDir, "active_urls.txt")
	err = utils.WriteFileLines(activeUrlsFile, activeURLs)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error writing active URLs to file: %v", err))
	}

	logFunc(fmt.Sprintf("✅ URL discovery and probing complete. Found %d unique URLs and %d active URLs.", len(uniqueURLs), len(activeURLs)), "")

	// After finding active URLs, run the parameter and pattern finder
	RunParameterAndPatternFinder(projectName, subtask, logFunc)
}

// runToolsOnFile executes a list of tools sequentially on a given input file, saving output to files.
func runToolsOnFile(projectName, inputFile string, tools []string, logFunc func(string, string)) error {
	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		return fmt.Errorf("error getting results directory: %v", err)
	}
	tempUrlsDir := filepath.Join(resultsDir, "urls", "temp")
	os.MkdirAll(tempUrlsDir, os.ModePerm)

	for _, toolName := range tools {
		outputFile := filepath.Join(tempUrlsDir, fmt.Sprintf("%s_%s_urls.txt", toolName, filepath.Base(inputFile)))

		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			logFunc("⚠️ GITHUB_TOKEN environment variable not set. Skipping github-endpoints.", "")
		}

		toolCmds := map[string]string{
			"gau":              fmt.Sprintf("cat %s | gau -t 5 | anew -q %s", inputFile, outputFile),
			"waybackurls":      fmt.Sprintf("cat %s | waybackurls | anew -q %s", inputFile, outputFile),
			"urlfinder":        fmt.Sprintf("urlfinder -list %s -all -silent -o %s", inputFile, outputFile),
			"hakrawler":        fmt.Sprintf("cat %s | hakrawler -timeout 5 -d 3 | anew -q %s", inputFile, outputFile),
	//		"katana":           fmt.Sprintf("katana -list %s -silent  -headless -d 6 -c 20 -jc -f qurl | anew -q %s", inputFile, outputFile),
			"gospider":         fmt.Sprintf(`gospider -S %s -t 100 -d 8 -c 10 | grep -Eo 'https?://[^ ]+' | sed 's/]$//' | anew -q %s`, inputFile, outputFile),
			"github-endpoints": fmt.Sprintf("while read -r line; do github-endpoints -d \"$line\" -raw -t %s -o github-endpoints_temp && cat github-endpoints_temp | anew -q %s && rm -f github-endpoints_temp; done < %s", githubToken, outputFile, inputFile),
			"cariddi":          fmt.Sprintf("cat %s | cariddi -plain | anew -q %s", inputFile, outputFile),
			"gourlex":          fmt.Sprintf("while read -r lines; do gourlex -t \"$lines\" -uO -s | anew -q %s; done < %s", outputFile, inputFile),
			"orwa":             fmt.Sprintf("bash ~/tools/orwa.sh %s | egrep 'http|https' | anew -q %s", inputFile, outputFile),
	//		"waymore":	    fmt.Sprintf("cat %s | waymore -mode U -p 5 -c ~/.config/waymore/config.yml -oU %s", inputFile, outputFile),
		}

		cmdStr, ok := toolCmds[toolName]
		if !ok {
			logFunc(fmt.Sprintf("⚠️ Unknown tool: %s", toolName), "")
			continue
		}

		logFunc(fmt.Sprintf("🔥 Running %s on %s...", toolName, inputFile), "")
		cmd := exec.Command("bash", "-c", cmdStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logFunc(fmt.Sprintf("⚠️ Error running %s on %s: %v. Output: %s", toolName, inputFile, err, string(output)), "")
			logFunc(fmt.Sprintf("✅ %s finished. Output file may be empty or incomplete.", toolName), "")
			continue
		}

		content, readErr := os.ReadFile(outputFile)
		if readErr != nil {
			logFunc(fmt.Sprintf("⚠️ Error reading file for line count %s: %v", outputFile, readErr), "")
		} else {
			lines := strings.Split(string(content), "\n")
			// Filter out empty lines before counting
			var validLines []string
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					validLines = append(validLines, line)
				}
			}
			logFunc(fmt.Sprintf("✅ %s finished. Found %d URLs. Results saved to %s", toolName, len(validLines), outputFile), "")
		}
	}

	return nil
}

func probeURLs(urls []string, logFunc func(string, string)) []string {
	var activeURLs []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	var probedCount int32
	totalURLs := len(urls)

	// List of user agents to rotate through
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
	}
	uaIndex := 0

	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirects
		},
	}

	concurrencyLimit := 200
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

			// Rotate user agent
			mu.Lock()
			ua := userAgents[uaIndex%len(userAgents)]
			uaIndex++
			mu.Unlock()

			req, err := http.NewRequest("GET", u, nil)
			if err != nil {
				atomic.AddInt32(&probedCount, 1)
				return
			}
			req.Header.Set("User-Agent", ua)

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt32(&probedCount, 1)
				return
			}
			defer resp.Body.Close()

			// Consider all status codes from 1xx to 5xx as active
			if resp.StatusCode >= 100 && resp.StatusCode < 600 {
				mu.Lock()
				activeURLs = append(activeURLs, u)
				mu.Unlock()
			}

			// Update progress
			newCount := atomic.AddInt32(&probedCount, 1)
			logFunc(fmt.Sprintf("\rProbed (%d/%d)", newCount, totalURLs), "")
		}(url)
	}

	wg.Wait()
	logFunc("\n", "") // Print a newline after probing is complete
	return activeURLs
}

func RunParameterAndPatternFinder(projectName string, subtask *tasks.Subtask, logFunc func(string, string)) {
	subtask.SetStatus(tasks.StatusRunning)
	defer subtask.SetStatus(tasks.StatusCompleted)

	logFunc("🔥 Starting parameter extraction and GF pattern matching...", "")

	resultsDir, err := utils.GetResultsDir(projectName)
	if err != nil {
		subtask.SetError(fmt.Sprintf("Error getting results directory: %v", err))
		logFunc(fmt.Sprintf("⚠️ Error getting results directory: %v", err), "")
		return
	}

	activeUrlsFile := filepath.Join(resultsDir, "urls", "active_urls.txt")
	if _, err := os.Stat(activeUrlsFile); os.IsNotExist(err) {
		logFunc("ℹ️ active_urls.txt not found, skipping parameter and pattern finding.", "")
		return
	}

	// 1. Extract parameters
	parameterFile := filepath.Join(resultsDir, "urls", "parameter.txt")
	logFunc("🔍 Extracting URLs with parameters...", "")
	// Command: cat active_urls.txt | grep "=" | grep -vE "\\.(woff|woff2|svg|json|js)$" > parameter.txt
	cmdStr := fmt.Sprintf("cat %s | grep \"=\" | grep -vE \"\\\\.(woff|woff2|svg|json|js)$ \" > %s", activeUrlsFile, parameterFile)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logFunc(fmt.Sprintf("⚠️ Error extracting parameters: %v. Output: %s", err, string(output)), "")
		// Continue to GF patterns even if this fails, maybe the file is empty.
	} else {
		logFunc(fmt.Sprintf("✅ Successfully extracted URLs with parameters to %s", parameterFile), "")
	}

	// 2. Run GF patterns
	patternsDir := filepath.Join(resultsDir, "patterns")
	os.MkdirAll(patternsDir, os.ModePerm)

	gfPatternsDir := filepath.Join(os.Getenv("HOME"), ".gf")
	files, err := os.ReadDir(gfPatternsDir)
	if err != nil {
		logFunc(fmt.Sprintf("⚠️ Error reading GF patterns directory (~/.gf): %v. Skipping GF scan.", err), "")
		return
	}

	logFunc(fmt.Sprintf("🏃 Running GF patterns from %s...", gfPatternsDir), "")

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			patternName := strings.TrimSuffix(file.Name(), ".json")
			logFunc(fmt.Sprintf("🎯 Running GF pattern: %s", patternName), "")

			outputFile := filepath.Join(patternsDir, fmt.Sprintf("%s.txt", patternName))
			// Command: gf <pattern> <urls_file> > <output_file>
			gfCmdStr := fmt.Sprintf("gf %s %s > %s", patternName, activeUrlsFile, outputFile)
			gfCmd := exec.Command("bash", "-c", gfCmdStr)
			gfOutput, err := gfCmd.CombinedOutput()
			if err != nil {
				logFunc(fmt.Sprintf("⚠️ Error running GF pattern %s: %v. Output: %s", patternName, err, string(gfOutput)), "")
				continue
			}
			logFunc(fmt.Sprintf("✅ Finished GF pattern: %s. Results saved to %s", patternName, outputFile), "")
		}
	}

	logFunc("✅ Parameter extraction and GF pattern matching complete.", "completed")
}
