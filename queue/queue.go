package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"hahacking/config"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Task struct {
	ID        int         `json:"id"`
	Method    string      `json:"method"`
	Params    interface{} `json:"params"`
	Result    interface{} `json:"result,omitempty"`
	Status    string      `json:"status"`
	Timestamp int64       `json:"timestamp"`
}

var (
	redisClient *redis.Client
	ctx         = context.Background()
	taskQueue   = "task_queue"
	wg          sync.WaitGroup
)

func Setup() {
	redisInfo := config.Get("redis")
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisInfo["host"].(string) + ":" + redisInfo["port"].(string),
		Password: redisInfo["password"].(string), // 密码为空
		DB:       int(redisInfo["db"].(float64)), // 使用默认数据库
	})
	// 启动工作池（例如3个worker）
	startWorkerPool(3)

	// 模拟添加任务
	tasks := []Task{
		{ID: 1, Method: "processData", Params: "test data", Status: "pending"},
		{ID: 2, Method: "calculateSum", Params: []int{1, 2, 3}, Status: "pending"},
	}

	for _, task := range tasks {
		if err := enqueueTask(task); err != nil {
			log.Printf("Failed to enqueue task %d: %v", task.ID, err)
		}
	}

	// 等待工作完成（实际应用中可根据需求调整）
	wg.Wait()
	log.Println("All tasks processed")
}
func Cancel() {
	wg.Done()
}

func enqueueTask(task Task) error {
	task.Timestamp = time.Now().Unix()
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return redisClient.RPush(ctx, taskQueue, data).Err()
}

func dequeueTask() (Task, bool) {
	result, err := redisClient.BLPop(ctx, 0, taskQueue).Result()
	if err != nil || len(result) < 2 {
		return Task{}, false
	}
	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		log.Printf("Unmarshal error: %v", err)
		return Task{}, false
	}
	return task, true
}

func executeTask(task Task) {
	defer wg.Done()

	// 模拟同步方法
	switch task.Method {
	case "processData":
		// 假设同步方法：processData(params string) string
		params, ok := task.Params.(string)
		if !ok {
			task.Status = "failed"
			return
		}
		result := processData(params)
		task.Result = result
		task.Status = "completed"
	case "calculateSum":
		// 假设同步方法：calculateSum(params []int) int
		params, ok := task.Params.([]int)
		if !ok {
			task.Status = "failed"
			return
		}
		result := calculateSum(params)
		task.Result = result
		task.Status = "completed"
	default:
		task.Status = "unknown"
	}

	// 更新任务状态到Redis
	data, _ := json.Marshal(task)
	redisClient.Set(ctx, fmt.Sprintf("task:%d", task.ID), data, 0)
}

func processData(params string) string {
	time.Sleep(5 * time.Second) // 模拟处理时间
	return "Processed: " + params
}

func calculateSum(params []int) int {
	time.Sleep(7 * time.Second) // 模拟处理时间
	sum := 0
	for _, v := range params {
		sum += v
	}
	return sum
}

func startWorkerPool(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					//收到退出信号
					return
				default:
					if task, ok := dequeueTask(); ok {
						executeTask(task)
					} else {
						time.Sleep(100 * time.Millisecond) // 避免忙等待
					}
				}
			}
		}()
	}
}
