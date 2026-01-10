package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setup() {
	Init()
}

func TestScheduler(t *testing.T) {
	setup()

	S.AddTask(&Task{
		ID:       "default", //注意这个ID很重要 - 清理时会用到
		Name:     "default",
		Schedule: "*/1 * * * * *", // 每1秒执行一次
		Type:     "default",
		Params: map[string]any{
			"name": "default",
			"born": time.Now(),
		},
	})

	tasks := S.GetTasks()
	assert.Equal(t, 1, len(tasks))
	t.Log(S.status)
	t.Log(S.taskIDs)
	t.Log(S.tasks)
	t.Log(S.executors)
	time.Sleep(1 * time.Second)
	tasks = S.GetTasks()
	assert.Equal(t, 0, len(tasks))

}

// 测试创建用户
func TestRunNow(t *testing.T) {
	setup()

	assert.Equal(t, 0, len(S.GetTasks()))

	S.AddTask(&Task{
		ID:       "test", //注意这个ID很重要 - 清理时会用到
		Name:     "test",
		Schedule: "*/2 * * * * *", // 每2秒执行一次
		Type:     "test",
		Params: map[string]any{
			"name": "test",
			"born": time.Now(),
		},
	})
	assert.Equal(t, 1, len(S.GetTasks()))
	S.RunNow("test")
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, len(S.GetTasks()))

}
