package scheduler

type Task struct {
	ID       string
	Name     string
	Schedule string                 `json:"schedule"` // Cron 表达式
	Params   map[string]interface{} `json:"params"`   // 任务参数
	Retry    int                    //LATER
	Type     string                 `json:"type"` // 任务类型 - 决定executor
}
