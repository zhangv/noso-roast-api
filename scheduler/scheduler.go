package scheduler

import (
	"fmt"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

var S *Scheduler

func Init() {
	S = NewScheduler()
	RegisterExecutors()
	S.Start()
	// defer S.Stop()
}

func RegisterExecutors() {
	S.RegisterExecutor("default", &DefaultTaskExecutor{})
	S.RegisterExecutor("test", &TestTaskExecutor{})
}

type Scheduler struct {
	c         *cron.Cron
	mu        sync.Mutex
	taskIDs   map[string]cron.EntryID
	tasks     map[string]*Task
	status    map[string]string
	executors map[string]Executor
}

func (s *Scheduler) RegisterExecutor(taskType string, executor Executor) {
	s.executors[taskType] = executor
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		c:         cron.New(cron.WithSeconds()),
		taskIDs:   make(map[string]cron.EntryID),
		tasks:     make(map[string]*Task),
		status:    make(map[string]string),
		executors: make(map[string]Executor),
	}
}

func (s *Scheduler) Start() {
	log.Println("Scheduler starting...")
	s.c.Start()
	log.Println("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.c.Stop()
}
func (s *Scheduler) GetTasks() map[string]*Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tasks
}

func (s *Scheduler) AddTask(t *Task) error {
	// s.mu.Lock()
	// defer s.mu.Unlock()

	executor, exists := s.executors[t.Type]
	if !exists {
		return fmt.Errorf("unknown task type: %s", t.Type)
	}
	id, err := s.c.AddFunc(t.Schedule, func() {
		log.Printf("Running task: %s", t.Name)
		if err := executor.Execute(t); err != nil {
			log.Printf("Failed to execute task %s: %v", t.Name, err)
		} else {
			log.Printf("Task %s executed successfully", t.Name)
			s.status[t.ID] = "success"
			s.RemoveTask(t.ID)
		}
	})
	if err != nil {
		return err
	}
	// 再加锁，操作map
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taskIDs[t.ID] = id
	s.tasks[t.ID] = t
	s.status[t.ID] = "running"
	return nil
}
func (s *Scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.taskIDs[taskID]; ok {
		s.c.Remove(id)
		delete(s.taskIDs, taskID)
	}
	delete(s.tasks, taskID)
	delete(s.status, taskID)
}

func (s *Scheduler) RunNow(taskID string) {
	s.mu.Lock()
	// defer s.mu.Unlock()
	t := s.tasks[taskID]
	executor, exists := s.executors[t.Type]
	s.mu.Unlock()
	if !exists {
		fmt.Errorf("unknown task type: %s", t.Type)
	}
	if err := executor.Execute(t); err != nil {
		log.Printf("Failed to execute task %s: %v", t.Name, err)
	} else {
		log.Printf("Task %s executed successfully", t.Name)
		s.status[t.ID] = "success"
		s.RemoveTask(t.ID)
	}
}

func (s *Scheduler) PauseTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.taskIDs[taskID]; ok {
		s.c.Remove(id)
		s.status[taskID] = "paused"
	}
}

func (s *Scheduler) ResumeTask(taskID string) error {
	if t, ok := s.tasks[taskID]; ok {
		return s.AddTask(t)
	}
	return nil
}

func (s *Scheduler) RetryTask(taskID string) {
	s.RunNow(taskID) // 简单实现，等同于立即执行
}

func (s *Scheduler) Status(taskID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status[taskID]
}
