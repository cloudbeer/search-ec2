package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(serviceManager interface{}) *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
}

// Check 健康检查 - 简化版本，直接返回 ok
func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
	}

	SuccessResponse(c, response)
}
