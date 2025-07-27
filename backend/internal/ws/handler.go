package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"project-saam/backend/internal/recon"
	"project-saam/backend/internal/tasks"
	"strings"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type WebSocketRequest struct {
	Action  string   `json:"action"`
	Project string   `json:"project"`
	Modules []string `json:"modules,omitempty"`
}

type WebSocketResponse struct {
	Log      string `json:"log,omitempty"`
	Progress string `json:"progress,omitempty"`
	Status   string `json:"status,omitempty"`
}

var connections = make(map[string]*websocket.Conn)
var mu sync.Mutex

func ReconStreamHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection attempt received.")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	var project string // Keep track of the project for this connection

	defer func() {
		mu.Lock()
		if project != "" {
			delete(connections, project)
		}
		mu.Unlock()
		conn.Close()
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		var req WebSocketRequest
		if err := json.Unmarshal(p, &req); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}
		log.Printf("Received WebSocket request: %+v", req) // Added logging

		project = req.Project
		mu.Lock()
		connections[req.Project] = conn
		mu.Unlock()

		task, exists := tasks.GetTask(req.Project)

		switch req.Action {
		case "start":
			log.Printf("Attempting to start recon for project: %s with modules: %v", req.Project, req.Modules) // Added logging
			if exists && task.Status == tasks.StatusRunning {
				sendMessage(req.Project, "A task is already running for this project.", string(task.Status))
				log.Printf("Task already running for project %s", req.Project) // Added logging
				continue
			}
			task = tasks.GetOrCreateTask(req.Project)
			task.SetStatus(tasks.StatusRunning)

			go func() {
				recon.RunRecon(req.Project, req.Modules, task, func(logMsg, status string) {
					// If it's a progress update, save it and send it.
					if strings.HasPrefix(logMsg, "\r") {
						task.SetProgress(logMsg)
						sendProgress(req.Project, logMsg)
						return
					}

					// For regular messages, write to the log and send the whole log.
					if logMsg != "" {
						task.WriteLog(logMsg + "\n")
					}
					if status != "" {
						task.SetStatus(tasks.TaskStatus(status))
					}
					sendMessage(req.Project, task.GetLog(), string(task.Status))
				})
				// Final status update
				task.SetStatus(tasks.StatusCompleted)
				task.SetProgress("") // Clear progress on completion
				sendMessage(req.Project, task.GetLog(), string(tasks.StatusCompleted))
				log.Printf("Reconnaissance for project %s finished or stopped.", req.Project)
			}()

		case "pause":
			if exists && task.Status == tasks.StatusRunning {
				task.PauseChan <- true
			}
		case "resume":
			if exists && task.Status == tasks.StatusPaused {
				task.PauseChan <- false
			}
		case "stop":
			if exists && (task.Status == tasks.StatusRunning || task.Status == tasks.StatusPaused) {
				task.StopChan <- true
			}
		case "status":
			if exists {
				sendMessage(req.Project, task.GetLog(), string(task.Status))
				if task.Progress != "" {
					sendProgress(req.Project, task.Progress)
				}
			} else {
				sendMessage(req.Project, "", "stopped")
			}
		}
	}
}

func sendMessage(project, logMessage, status string) {
	mu.Lock()
	defer mu.Unlock()

	if conn, ok := connections[project]; ok {
		res := WebSocketResponse{Log: logMessage, Status: status}
		if err := conn.WriteJSON(res); err != nil {
			log.Println("Write error:", err)
		}
	}
}

func sendProgress(project, progress string) {
	mu.Lock()
	defer mu.Unlock()

	if conn, ok := connections[project]; ok {
		res := WebSocketResponse{Progress: progress}
		if err := conn.WriteJSON(res); err != nil {
			log.Println("Write error:", err)
		}
	}
}
