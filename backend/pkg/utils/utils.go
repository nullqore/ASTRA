package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Dir(dir), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root")
		}
		dir = parent
	}
}

func GetResultsDir(projectName string) (string, error) {
	resultsDir := os.Getenv("SAAM_RESULTS_DIR")
	if resultsDir != "" {
		return filepath.Join(resultsDir, projectName), nil
	}
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, "results", projectName), nil
}

func ReadFileLines(path string) ([]string, error) {
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

func WriteFileLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func CombineFiles(outputFile string, files ...string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	for _, file := range files {
		inFile, err := os.Open(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			fmt.Fprintln(writer, scanner.Text())
		}
		inFile.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func RunCommand(command string, logFunc func(string, string)) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logFunc(fmt.Sprintf("Error running command: %s", err), "error")
	}
	return err
}
