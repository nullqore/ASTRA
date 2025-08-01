package handlers

import (
        "encoding/json"
        "fmt"
        "net/http"
        "os"
        "path/filepath"
        "project-saam/backend/pkg/utils"
        "strings"
        "time"
)

// --- Structs ---
type CreateProjectRequest struct {
        ProjectName string `json:"projectName"`
}

type AddTargetRequest struct {
        Target string `json:"target"`
        Type   string `json:"type"`
}

type Project struct {
        Name       string    `json:"name"`
        CreatedAt  time.Time `json:"created_at"`
        Domains    []string  `json:"domains"`
        Wildcards  []string  `json:"wildcards"`
        OutOfScope []string  `json:"out_of_scope"`
}

type Module struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        Locked      bool   `json:"locked"`
}

type ProjectStats struct {
	Domains    int `json:"domains"`
	Wildcards  int `json:"wildcards"`
	Subdomains int `json:"subdomains"`
	URLs       int `json:"urls"`
	JSURLs     int `json:"js_urls"`
}

// --- Utility Functions ---
// findProjectRoot walks up from the current working directory to find the directory containing go.mod
func findProjectRoot() (string, error) {
        dir, err := os.Getwd()
        if err != nil {
                return "", err
        }
        for {
                // Check for go.mod to identify the backend directory
                if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
                        // The project root is the parent of the backend directory
                        return filepath.Dir(dir), nil
                }
                parent := filepath.Dir(dir)
                if parent == dir {
                        return "", fmt.Errorf("could not find project root (go.mod not found)")
                }
                dir = parent
        }
}

func getResultsDir() (string, error) {
        resultsDir := os.Getenv("SAAM_RESULTS_DIR")
        if resultsDir != "" {
                return resultsDir, nil
        }
        projectRoot, err := findProjectRoot()
        if err != nil {
                return "", err
        }
        return filepath.Join(projectRoot, "results"), nil
}

// --- Handlers ---
func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
        var req CreateProjectRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Error decoding request body", http.StatusBadRequest)
                return
        }
        resultsDir, err := getResultsDir()
        if err != nil {
                http.Error(w, fmt.Sprintf("Could not determine results directory: %v", err), http.StatusInternalServerError)
                return
        }
        projectPath := filepath.Join(resultsDir, req.ProjectName)

        if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
                project, err := getProjectByName(req.ProjectName)
                if err != nil {
                        http.Error(w, "Project already exists but could not retrieve details", http.StatusInternalServerError)
                        return
                }
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(project)
                return
        }

        projectScopePath := filepath.Join(projectPath, "scope")
        if err := os.MkdirAll(projectScopePath, os.ModePerm); err != nil {
                http.Error(w, fmt.Sprintf("Failed to create project directory: %v", err), http.StatusInternalServerError)
                return
        }

        utils.WriteFileLines(filepath.Join(projectScopePath, "domain.txt"), []string{})
        utils.WriteFileLines(filepath.Join(projectScopePath, "wildcard.txt"), []string{})
        utils.WriteFileLines(filepath.Join(projectScopePath, "out-of-scope.txt"), []string{})

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(Project{Name: req.ProjectName, CreatedAt: time.Now()})
}

func GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
        resultsDir, _ := getResultsDir()
        files, _ := os.ReadDir(resultsDir)
        var projects []Project
        for _, file := range files {
                if file.IsDir() {
                        info, _ := file.Info()
                        projects = append(projects, Project{Name: file.Name(), CreatedAt: info.ModTime()})
                }
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(projects)
}

func getProjectByName(projectName string) (*Project, error) {
        resultsDir, _ := getResultsDir()
        projectPath := filepath.Join(resultsDir, projectName)
        info, err := os.Stat(projectPath)
        if err != nil {
                return nil, err
        }
        scopePath := filepath.Join(projectPath, "scope")
        domains, _ := utils.ReadFileLines(filepath.Join(scopePath, "domain.txt"))
        wildcards, _ := utils.ReadFileLines(filepath.Join(scopePath, "wildcard.txt"))
        outOfScope, _ := utils.ReadFileLines(filepath.Join(scopePath, "out-of-scope.txt"))

        if domains == nil {
                domains = []string{}
        }
        if wildcards == nil {
                wildcards = []string{}
        }
        if outOfScope == nil {
                outOfScope = []string{}
        }

        project := &Project{
                Name:       info.Name(),
                CreatedAt:  info.ModTime(),
                Domains:    domains,
                Wildcards:  wildcards,
                OutOfScope: outOfScope,
        }
        return project, nil
}

func GetProjectByNameHandler(w http.ResponseWriter, r *http.Request) {
        projectName := strings.TrimPrefix(r.URL.Path, "/api/projects/")
        project, err := getProjectByName(projectName)
        if err != nil {
                http.Error(w, "Project not found", http.StatusNotFound)
                return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(project)
}

func AddTargetHandler(w http.ResponseWriter, r *http.Request) {
        projectName := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/targets")
        var req AddTargetRequest
        json.NewDecoder(r.Body).Decode(&req)
        resultsDir, _ := getResultsDir()
        filePath := filepath.Join(resultsDir, projectName, "scope", req.Type+".txt")
        existingLines, _ := utils.ReadFileLines(filePath)

        for _, line := range existingLines {
                if strings.TrimSpace(line) == strings.TrimSpace(req.Target) {
                        w.Header().Set("Content-Type", "application/json")
                        w.WriteHeader(http.StatusConflict)
                        json.NewEncoder(w).Encode(map[string]string{"error": "Duplicate target"})
                        return
                }
        }

        file, _ := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
        defer file.Close()
        file.WriteString(strings.TrimSpace(req.Target) + "\n")

        updatedProject, err := getProjectByName(projectName)
        if err != nil {
                http.Error(w, "Could not retrieve updated project", http.StatusInternalServerError)
                return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(updatedProject)
}

func RemoveTargetHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodDelete {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }
        path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
        projectName := strings.Split(path, "/")[0]
        targetToRemove := r.URL.Query().Get("target")
        targetType := r.URL.Query().Get("type")

        resultsDir, _ := getResultsDir()
        filePath := filepath.Join(resultsDir, projectName, "scope", targetType+".txt")
        lines, _ := utils.ReadFileLines(filePath)

        var newLines []string
        found := false
        for _, line := range lines {
                if strings.TrimSpace(line) != strings.TrimSpace(targetToRemove) {
                        newLines = append(newLines, line)
                } else {
                        found = true
                }
        }

        if found {
                utils.WriteFileLines(filePath, newLines)
        }

        updatedProject, err := getProjectByName(projectName)
        if err != nil {
                http.Error(w, "Could not retrieve updated project", http.StatusInternalServerError)
                return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(updatedProject)
}

func GetModulesHandler(w http.ResponseWriter, r *http.Request) {
        modules := []Module{
                {Name: "Recon", Description: "Perform reconnaissance on targets.", Locked: false},
                {Name: "Monitor", Description: "Monitor targets for changes.", Locked: true},
                {Name: "Mindmap", Description: "Visualize project data.", Locked: true},
                {Name: "Empty", Description: "A placeholder module.", Locked: true},
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(modules)
}

func GetProjectStatsHandler(w http.ResponseWriter, r *http.Request) {
	projectName := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	projectName = strings.TrimSuffix(projectName, "/stats")

	resultsDir, err := getResultsDir()
	if err != nil {
		http.Error(w, "Could not determine results directory", http.StatusInternalServerError)
		return
	}
	projectPath := filepath.Join(resultsDir, projectName)

	domains, _ := utils.ReadFileLines(filepath.Join(projectPath, "scope", "domain.txt"))
	wildcards, _ := utils.ReadFileLines(filepath.Join(projectPath, "scope", "wildcard.txt"))
	subdomains, _ := utils.ReadFileLines(filepath.Join(projectPath, "active", "active-subs.txt"))
	urls, _ := utils.ReadFileLines(filepath.Join(projectPath, "urls", "active_urls.txt"))
	jsUrls, _ := utils.ReadFileLines(filepath.Join(projectPath, "urls", "active-js-urls.txt"))

	stats := ProjectStats{
		Domains:    len(domains),
		Wildcards:  len(wildcards),
		Subdomains: len(subdomains),
		URLs:       len(urls),
		JSURLs:     len(jsUrls),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
