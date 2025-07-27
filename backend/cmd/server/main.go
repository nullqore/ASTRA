package main

import (
        "log"
        "net/http"
        "strings"
        "project-saam/backend/internal/handlers"
        "project-saam/backend/internal/ws"
        "github.com/joho/godotenv"
        "github.com/rs/cors"
)

func projectMux(w http.ResponseWriter, r *http.Request) {
        path := strings.TrimPrefix(r.URL.Path, "/api/projects/")

        if strings.Contains(path, "/targets") {
                if r.Method == http.MethodPost {
                        handlers.AddTargetHandler(w, r)
                        return
                }
                if r.Method == http.MethodDelete {
                        handlers.RemoveTargetHandler(w, r)
                        return
                }
        }

        handlers.GetProjectByNameHandler(w, r)
}

func main() {
        err := godotenv.Load()
        if err != nil {
                log.Println("Note: .env file not found, proceeding with system environment variables")
        }
        mux := http.NewServeMux()

        // Serve the frontend files
        fs := http.FileServer(http.Dir("./frontend"))
        mux.Handle("/", fs)

        // Setup API routes
        mux.HandleFunc("/api/create-project", handlers.CreateProjectHandler)
        mux.HandleFunc("/api/projects", handlers.GetProjectsHandler)
        mux.HandleFunc("/api/projects/", projectMux)
        mux.HandleFunc("/api/modules", handlers.GetModulesHandler)
        mux.HandleFunc("/ws", ws.ReconStreamHandler)

        // Configure CORS options
        c := cors.New(cors.Options{
                AllowedOrigins:   []string{"*"}, // Allow any origin
                AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
                AllowedHeaders:   []string{"*"},
                AllowCredentials: true,
        })

        // Wrap your main router with the CORS middleware
        handler := c.Handler(mux)

        // Use the new handler to start the server
        log.Println("Server starting on :8080")
        log.Fatal(http.ListenAndServe(":8080", handler))
}
