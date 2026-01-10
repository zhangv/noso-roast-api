package scheduler

import (
	"log"
	"time"
)

type Executor interface {
	Execute(task *Task) error
}

type DefaultTaskExecutor struct {
}

func (e *DefaultTaskExecutor) Execute(task *Task) error {
	log.Println("Executing task ID:", task.ID)
	log.Println("Executing task Name:", task.Name)
	log.Println("Executing task Schedule:", task.Schedule)
	log.Println("Executing task Type:", task.Type)
	log.Println("Executing task Params:", task.Params)
	return nil
}

type TestTaskExecutor struct {
}

func (e *TestTaskExecutor) Execute(task *Task) error {
	log.Printf("Archiving logs for task: %s, Params: %+v", task.Name, task.Params)
	time.Sleep(1 * time.Second)
	log.Printf("Log archival completed for task: %s", task.Name)

	return nil
}
