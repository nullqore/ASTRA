package tasks

import (
	"bytes"
	"sync"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusPaused    TaskStatus = "paused"
	StatusStopped   TaskStatus = "stopped"
	StatusCompleted TaskStatus = "completed"
	StatusError     TaskStatus = "error"
)

type Subtask struct {
	Name   string
	Status TaskStatus
	Error  string
	mu     sync.RWMutex
}

func (s *Subtask) SetStatus(status TaskStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

func (s *Subtask) SetError(err string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = StatusError
	s.Error = err
}

type Task struct {
	ProjectName string
	Status      TaskStatus
	LogBuffer   *bytes.Buffer
	Progress    string
	PauseChan   chan bool
	StopChan    chan bool
	Subtasks    map[string]*Subtask
	mu          sync.RWMutex
}

var (
	tasks = make(map[string]*Task)
	mu    sync.RWMutex
)

func GetOrCreateTask(projectName string) *Task {
	mu.Lock()
	defer mu.Unlock()

	if task, exists := tasks[projectName]; exists {
		if task.Status == StatusStopped || task.Status == StatusCompleted || task.Status == StatusError {
			task.LogBuffer.Reset()
			task.PauseChan = make(chan bool)
			task.StopChan = make(chan bool)
			task.Subtasks = make(map[string]*Subtask)
		}
		return task
	}

	task := &Task{
		ProjectName: projectName,
		Status:      StatusStopped,
		LogBuffer:   new(bytes.Buffer),
		PauseChan:   make(chan bool),
		StopChan:    make(chan bool),
		Subtasks:    make(map[string]*Subtask),
	}
	tasks[projectName] = task
	return task
}

func GetTask(projectName string) (*Task, bool) {
	mu.RLock()
	defer mu.RUnlock()
	task, exists := tasks[projectName]
	return task, exists
}

func (t *Task) SetStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

func (t *Task) WriteLog(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.LogBuffer.WriteString(message + "\n")
}

func (t *Task) GetLog() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.LogBuffer.String()
}

func (t *Task) SetProgress(progress string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Progress = progress
}

func (t *Task) AddSubtask(key string, subtask *Subtask) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Subtasks[key] = subtask
}

func (t *Task) GetSubtask(key string) *Subtask {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Subtasks[key]
}
