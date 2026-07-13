package handlers

import (
	"net/http"
	"strconv"

	"github.com/aistudio/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// LogHandler handles log query operations.
type LogHandler struct {
	svc *service.LogService
}

// NewLogHandler creates a new LogHandler.
func NewLogHandler(svc *service.LogService) *LogHandler {
	return &LogHandler{svc: svc}
}

// Query returns filtered and paginated log entries.
// GET /api/logs
// Query parameters:
//   - level: DEBUG|INFO|WARN|ERROR
//   - source: filter by source
//   - taskId: filter by task ID
//   - keyword: search in message
//   - start: RFC3339 start time
//   - end: RFC3339 end time
//   - page: page number (default 1)
//   - size: page size (default 20, max 100)
func (h *LogHandler) Query(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	q := service.LogQuery{
		Level:   service.LogLevel(c.Query("level")),
		Source:  c.Query("source"),
		TaskID:  c.Query("taskId"),
		Keyword: c.Query("keyword"),
		Start:   c.Query("start"),
		End:     c.Query("end"),
		Page:    page,
		Size:    size,
	}

	result, err := h.svc.Query(q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

// FetchTaskLogs returns logs for a specific task.
// GET /api/tasks/:taskId/logs
func (h *LogHandler) FetchTaskLogs(c *gin.Context) {
	taskID := c.Param("taskId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "100"))

	q := service.LogQuery{
		TaskID: taskID,
		Page:   page,
		Size:   size,
	}

	result, err := h.svc.Query(q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}