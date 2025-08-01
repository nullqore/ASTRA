package routes

import (
	"net/http"
	"project-saam/backend/internal/handlers"
	"project-saam/backend/internal/ws"
	"strings"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	createProjectHandler := corsMiddleware(http.HandlerFunc(handlers.CreateProjectHandler))
	getModulesHandler := corsMiddleware(http.HandlerFunc(handlers.GetModulesHandler))
	
	mux.Handle("/api/create-project", createProjectHandler)
	mux.Handle("/api/modules", getModulesHandler)

	projectsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/targets") {
			switch r.Method {
			case http.MethodPost:
				handlers.AddTargetHandler(w, r)
			case http.MethodDelete:
				handlers.RemoveTargetHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			if r.Method == http.MethodGet {
				if path == "/api/projects/" {
					handlers.GetProjectsHandler(w, r)
				} else {
					handlers.GetProjectByNameHandler(w, r)
				}
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	mux.Handle("/api/projects/", corsMiddleware(projectsHandler))
	mux.HandleFunc("/ws", ws.ReconStreamHandler)

	return mux
}
