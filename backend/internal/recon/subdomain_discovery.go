package recon

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"project-saam/backend/pkg/utils"
)

// runSubdomainDiscovery orchestrates the entire subdomain discovery process sequentially for each wildcard.
func (r *ReconOrchestrator) runSubdomainDiscovery() error {
	r.Log("🚀 Starting advanced subdomain discovery...", "")

	wildcardDomains, err := r.getWildcardDomains()
	if err != nil {
		r.Log(fmt.Sprintf("⚠️  Could not read wildcard domains: %v. Skipping subdomain discovery.", err), "")
		return nil // Not a fatal error
	}
	if len(wildcardDomains) == 0 {
		r.Log("ℹ️  No wildcard domains found in scope. Skipping subdomain discovery.", "")
		return nil
	}

	// Create the main 'subs' directory for the final results.
	subsDir := filepath.Join(r.ResultsDir, "subs")
	fmt.Printf("DEBUG: Subdomains output directory: %s\n", subsDir)
	if err := os.MkdirAll(subsDir, 0755); err != nil {
		return fmt.Errorf("failed to create subs directory: %w", err)
	}

	// Create a single temp directory for all intermediate tool outputs.
	tempDir := filepath.Join(subsDir, "temp")
	fmt.Printf("DEBUG: Temporary directory for tools: %s\n", tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up all temp files at the very end.

	r.Log(fmt.Sprintf("ℹ️  Found %d wildcard domains to process sequentially.", len(wildcardDomains)), "")

	// Process each domain sequentially.
	for i, domain := range wildcardDomains {
		r.Log(fmt.Sprintf("\n--- Processing wildcard %d/%d: %s ---", i+1, len(wildcardDomains), domain), "")

		// Run all discovery tools for the current domain.
		r.runToolsForDomain(domain, tempDir)

		// After all tools run for the domain, merge its specific results.
		finalOutputFile := filepath.Join(subsDir, fmt.Sprintf("%s_subs.txt", domain))
		fmt.Printf("DEBUG: Final output file for %s: %s\n", domain, finalOutputFile)
		if err := r.mergeDomainResults(domain, tempDir, finalOutputFile); err != nil {
			r.Log(fmt.Sprintf("❌ Error merging results for %s: %v", domain, err), "")
			// Continue to the next domain even if one fails.
		}
		r.Log(fmt.Sprintf("--- Finished processing wildcard: %s ---", domain), "")
	}

	r.Log("\n✅ Advanced subdomain discovery complete for all wildcards.", "")
	return nil
}

// runToolsForDomain executes all subdomain discovery tools for a single domain sequentially.
func (r *ReconOrchestrator) runToolsForDomain(domain string, tempDir string) {
	tools := []struct {
		Name   string
		Runner func(string, string)
	}{
		{Name: "subfinder", Runner: r.runSubfinder},
		{Name: "assetfinder", Runner: r.runAssetfinder},
		{Name: "chaos", Runner: r.runChaos},
		{Name: "findomain", Runner: r.runFindomain},
		{Name: "github-subdomains", Runner: r.runGithubSubdomains},
		{Name: "whoisxmlapi", Runner: r.runWhoisxmlapi},
		{Name: "crtsh", Runner: r.runCrtsh},
		{Name: "waybackurls", Runner: r.runWaybackurls},
		{Name: "puredns", Runner: r.runPuredns},
		{Name: "subdominator", Runner: r.runSubdominator},
	}

	for _, tool := range tools {
		select {
		case <-r.Ctx.Done():
			r.Log(fmt.Sprintf("⏹️  Reconnaissance stopped, skipping remaining tools for %s.", domain), "")
			return
		default:
		}
		outputFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s.txt", tool.Name, domain))
		tool.Runner(domain, outputFile)
	}
}

// Tool-specific runner functions
func (r *ReconOrchestrator) runSubfinder(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running subfinder for %s...", domain), "")
	cmd := exec.CommandContext(r.Ctx, "subfinder", "-d", domain, "-all", "-recursive", "-o", outputFile)
	r.executeCommand(cmd, "subfinder", domain)
}

func (r *ReconOrchestrator) runAssetfinder(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running assetfinder for %s...", domain), "")
	cmdStr := fmt.Sprintf("assetfinder --subs-only %s > %s", domain, outputFile)
	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	r.executeCommand(cmd, "assetfinder", domain, outputFile)
}

func (r *ReconOrchestrator) runChaos(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running chaos for %s...", domain), "")
	chaosKey := os.Getenv("CHAOS_KEY")
	if chaosKey == "" {
		r.Log("⚠️ CHAOS_KEY environment variable not set. Skipping chaos.", "")
		return
	}
	cmd := exec.CommandContext(r.Ctx, "chaos", "-d", domain, "-key", chaosKey, "-o", outputFile, "-silent")
	r.executeCommand(cmd, "chaos", domain)
}

func (r *ReconOrchestrator) runFindomain(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running findomain for %s...", domain), "")
	cmd := exec.CommandContext(r.Ctx, "findomain", "-t", domain, "-q", "-r", "--unique-output", outputFile)
	r.executeCommand(cmd, "findomain", domain)
}

func (r *ReconOrchestrator) runGithubSubdomains(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running github-subdomains for %s...", domain), "")
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		r.Log("⚠️ GITHUB_TOKEN environment variable not set. Skipping github-subdomains.", "")
		return
	}
	cmdStr := fmt.Sprintf("github-subdomains -t %s -d %s -o %s", githubToken, domain, outputFile)
	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	r.executeCommand(cmd, "github-subdomains", domain, outputFile)
}

func (r *ReconOrchestrator) runWhoisxmlapi(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running WhoisXMLAPI for %s...", domain), "")
	apiKey := os.Getenv("WHOISXML_API_KEY")
	if apiKey == "" {
		r.Log("⚠️ WHOISXML_API_KEY environment variable not set. Skipping WhoisXMLAPI.", "")
		return
	}
	cmdStr := fmt.Sprintf(`curl -s "https://subdomains.whoisxmlapi.com/api/v1?apiKey=%s&domainName=%s" | jq -r '.result.records[].domain' | sort -u > %s`, apiKey, domain, outputFile)
	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	r.executeCommand(cmd, "whoisxmlapi", domain)
}

func (r *ReconOrchestrator) runCrtsh(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running crt.sh for %s...", domain), "")
	urlPart := `curl -s "https://crt.sh/?q=%25.`
	cmdStr := urlPart + domain + `&output=json" | jq -r '.[].name_value' | sed 's/\*\.//g' | sort -u > ` + outputFile
	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	r.executeCommand(cmd, "crtsh", domain, outputFile)
}

func (r *ReconOrchestrator) runWaybackurls(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running Waybackurls for %s...", domain), "")
	cmdStr := fmt.Sprintf(`curl -s "http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=text&fl=original&collapse=urlkey" | sort | sed -e 's_https*://__' -e "s/\/.*//" -e 's/:.*//' -e 's/^www\.//' | sort -u > %s`, domain, outputFile)
	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	r.executeCommand(cmd, "waybackurls", domain)
}

func (r *ReconOrchestrator) runPuredns(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running puredns for %s...", domain), "")
	cmd := exec.CommandContext(r.Ctx, "puredns", "bruteforce", "~/tools/wordlists/sub-fuzzer.txt", domain, "--resolvers", "~/tools/wordlists/resolvers.txt", "-w", outputFile)
	r.executeCommand(cmd, "puredns", domain)
}

func (r *ReconOrchestrator) runSubdominator(domain, outputFile string) {
	r.Log(fmt.Sprintf("🔥 Running subdominator for %s...", domain), "")
	cmd := exec.CommandContext(r.Ctx, "subdominator", "-d", domain, "-o", outputFile)
	r.executeCommand(cmd, "subdominator", domain)
}

func (r *ReconOrchestrator) executeCommand(cmd *exec.Cmd, toolName, domain string, outputFile ...string) {
	// If an output file is specified, it means the command will redirect stdout.
	// In this case, we just need to run the command and let the shell handle the redirection.
	if len(outputFile) > 0 {
		if err := cmd.Run(); err != nil {
			if r.Ctx.Err() == context.Canceled {
				r.Log(fmt.Sprintf("⏹️  %s stopped by user for %s.", toolName, domain), "")
			} else {
				r.Log(fmt.Sprintf("❌ Error running %s for %s: %v", toolName, domain, err), "")
			}
			return
		}
	} else {
		// If no output file is specified, it means the tool supports the -o flag.
		// We execute the command and it will write to the file specified in its arguments.
		if err := cmd.Run(); err != nil {
			if r.Ctx.Err() == context.Canceled {
				r.Log(fmt.Sprintf("⏹️  %s stopped by user for %s.", toolName, domain), "")
			} else {
				r.Log(fmt.Sprintf("❌ Error running %s for %s: %v", toolName, domain, err), "")
			}
			return
		}
	}

	// Now, read the output file to count the results.
	var finalOutputFile string
	if len(outputFile) > 0 {
		finalOutputFile = outputFile[0]
	} else {
		// This assumes the output file is the last argument in the command.
		// This might not be robust for all commands, but it works for the current set.
		finalOutputFile = cmd.Args[len(cmd.Args)-1]
	}

	content, readErr := ioutil.ReadFile(finalOutputFile)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			r.Log(fmt.Sprintf("✅ %s finished for %s. Found 0 subdomains.", toolName, domain), "")
			return
		}
		r.Log(fmt.Sprintf("⚠️  Could not read output file for %s on %s: %v", toolName, domain, readErr), "")
		return
	}

	lines := strings.Split(string(content), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	r.Log(fmt.Sprintf("✅ %s finished for %s. Found %d subdomains.", toolName, domain, count), "")
}

// mergeDomainResults merges all temporary tool outputs for a specific domain,
// compares them with existing results, and saves the updated unique list.
func (r *ReconOrchestrator) mergeDomainResults(domain, tempDir, finalOutputFile string) error {
	r.Log(fmt.Sprintf("🔬 Applying advanced filtering and sorting for %s...", domain), "")

	// 1. Get existing subdomains to compare against
	var existingSubdomains = make(map[string]struct{})
	if _, err := os.Stat(finalOutputFile); !os.IsNotExist(err) {
		content, readErr := ioutil.ReadFile(finalOutputFile)
		if readErr != nil {
			r.Log(fmt.Sprintf("⚠️ Could not read existing results file for %s: %v. It will be overwritten.", domain, readErr), "")
		} else {
			for _, line := range strings.Split(string(content), "\n") {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					existingSubdomains[trimmed] = struct{}{}
				}
			}
		}
	}
	r.Log(fmt.Sprintf("ℹ️ Found %d existing subdomains for %s.", len(existingSubdomains), domain), "")

	// 2. Use a shell pipeline to process the new subdomain files.
	// This is faster and less memory-intensive than doing it in Go.
	escapedDomain := strings.ReplaceAll(domain, ".", "\\.")
	// Create a temporary file for the new, unique results from this run.
	tempFinalFile := finalOutputFile + ".tmp"
	defer os.Remove(tempFinalFile) // Ensure cleanup

	cmdStr := fmt.Sprintf(
		`cat %s/*_%s.txt 2>/dev/null | submore -d %s | grep -Ev '\*' | sed -E 's/^[[:space:]]* //g' | awk '{print tolower($0)}' | grep -E "\\b[[:alnum:]._-]*\\.%s$" | sort -u > %s`,
		tempDir, domain, domain, escapedDomain, tempFinalFile,
	)

	cmd := exec.CommandContext(r.Ctx, "bash", "-c", cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("subdomain filtering pipeline failed for %s: %w", domain, err)
	}

	// 3. Read the new results and find the truly new subdomains
	newContent, err := ioutil.ReadFile(tempFinalFile)
	if err != nil {
		if os.IsNotExist(err) {
			r.Log(fmt.Sprintf("✨ Filtering complete. No new subdomains found for %s this run.", domain), "")
			return nil
		}
		return fmt.Errorf("could not read temporary results file for %s: %w", domain, err)
	}

	newlyFoundCount := 0
	for _, line := range strings.Split(string(newContent), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			if _, exists := existingSubdomains[trimmed]; !exists {
				newlyFoundCount++
				// Add to the map to ensure the final list is unique
				existingSubdomains[trimmed] = struct{}{}
			}
		}
	}

	// 4. Write the combined, unique, and sorted list back to the final file.
	allUniqueSubdomains := make([]string, 0, len(existingSubdomains))
	for sub := range existingSubdomains {
		allUniqueSubdomains = append(allUniqueSubdomains, sub)
	}
	// The list is already unique, now sort it for consistency.
	// sort.Strings(allUniqueSubdomains) // The pipeline already sorts, so this is optional but good practice.

	if err := utils.WriteFileLines(finalOutputFile, allUniqueSubdomains); err != nil {
		return fmt.Errorf("failed to write final updated subdomain list for %s: %w", domain, err)
	}

	if newlyFoundCount > 0 {
		r.Log(fmt.Sprintf("🎉 Found %d brand new subdomains for %s! Total is now %d.", newlyFoundCount, domain, len(allUniqueSubdomains)), "")
	} else {
		r.Log(fmt.Sprintf("✨ Filtering complete. No new subdomains found for %s. Total remains %d.", domain, len(allUniqueSubdomains)), "")
	}

	return nil
}

// getWildcardDomains reads the wildcard domains from the project's scope.
func (r *ReconOrchestrator) getWildcardDomains() ([]string, error) {
	wildcardFile := filepath.Join(r.ResultsDir, "scope", "wildcard.txt")
	if _, err := os.Stat(wildcardFile); os.IsNotExist(err) {
		return []string{}, nil
	}
	content, err := utils.ReadFileLines(wildcardFile)
	if err != nil {
		return nil, err
	}
	var domains []string
	for _, line := range content {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			domains = append(domains, strings.TrimPrefix(trimmed, "*."))
		}
	}
	return domains, nil
}