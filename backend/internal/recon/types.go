package recon

import (
	"bufio"
	"os"
	"path/filepath"
)

// ReconStatus is used to report the status of a running tool.
type ReconStatus struct {
	Tool   string `json:"tool"`
	Status string `json:"status"`
	Count  int    `json:"count,omitempty"`
}

// --- Utility Functions (shared within the recon package) ---

func getResultsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Assuming the executable is in a `cmd` directory, this goes up to the project root
	return filepath.Join(filepath.Dir(wd), "results"), nil
}

func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
