package task

import (
	"hahacking/middleware"
	"hahacking/scheduler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/task")
	{
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.POST("add", Add)
		userAPI.POST("runNow/:id", RunNow)
		userAPI.POST("pause/:id", Pause)
		userAPI.POST("resume/:id", Resume)
	}
}

func Add(c *gin.Context) {
	var req struct {
		ID       string                 `json:"id"`
		Name     string                 `json:"name"`
		Schedule string                 `json:"schedule"`
		Type     string                 `json:"type"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task := &scheduler.Task{
		ID:       req.ID,
		Name:     req.Name,
		Schedule: req.Schedule,
		Type:     req.Type,
		Params:   req.Params,
	}
	err := scheduler.S.AddTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task added"})
}

func RunNow(c *gin.Context) {
	id := c.Param("id")
	scheduler.S.RunNow(id)
	c.JSON(http.StatusOK, gin.H{"message": "task executed now"})
}

func Pause(c *gin.Context) {
	id := c.Param("id")
	scheduler.S.PauseTask(id)
	c.JSON(http.StatusOK, gin.H{"message": "task paused"})
}

func Resume(c *gin.Context) {
	id := c.Param("id")
	err := scheduler.S.ResumeTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task resumed"})
}
