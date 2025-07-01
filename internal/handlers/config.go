package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"search-ec2/internal/config"
	"search-ec2/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	serviceManager *services.ServiceManager
}

// NewConfigHandler 创建配置管理处理器
func NewConfigHandler(serviceManager *services.ServiceManager) *ConfigHandler {
	return &ConfigHandler{
		serviceManager: serviceManager,
	}
}

// GetFunctionSchema 获取 Function Calling Schema
func (h *ConfigHandler) GetFunctionSchema(c *gin.Context) {
	if config.FunctionSchema == nil {
		InternalErrorResponse(c, "Function schema not loaded")
		return
	}

	SuccessResponse(c, config.FunctionSchema)
}

// UpdateFunctionSchema 更新 Function Calling Schema
func (h *ConfigHandler) UpdateFunctionSchema(c *gin.Context) {
	var newSchema config.FunctionCallingSchema
	if err := c.ShouldBindJSON(&newSchema); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid schema: %v", err))
		return
	}

	// 验证 schema 格式
	if newSchema.FunctionName == "" {
		BadRequestResponse(c, "Function name is required")
		return
	}

	if newSchema.Parameters == nil {
		BadRequestResponse(c, "Parameters are required")
		return
	}

	// 保存到文件
	schemaPath := "config/function_calling_schema.json"
	data, err := json.MarshalIndent(newSchema, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal schema: %v", err)
		InternalErrorResponse(c, "Failed to process schema")
		return
	}

	if err := os.WriteFile(schemaPath, data, 0644); err != nil {
		logrus.Errorf("Failed to write schema file: %v", err)
		InternalErrorResponse(c, "Failed to save schema")
		return
	}

	// 更新内存中的配置
	config.FunctionSchema = &newSchema

	logrus.Infof("Function calling schema updated successfully")

	response := map[string]interface{}{
		"message":       "Schema updated successfully",
		"function_name": newSchema.FunctionName,
		"updated_at":   "now",
	}

	SuccessResponse(c, response)
}

// GetVariantPrompt 获取变体生成提示词
func (h *ConfigHandler) GetVariantPrompt(c *gin.Context) {
	if config.VariantPromptTemplate == "" {
		InternalErrorResponse(c, "Variant prompt template not loaded")
		return
	}

	response := map[string]interface{}{
		"prompt": config.VariantPromptTemplate,
	}

	SuccessResponse(c, response)
}

// UpdateVariantPrompt 更新变体生成提示词
func (h *ConfigHandler) UpdateVariantPrompt(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if len(req.Prompt) < 10 {
		BadRequestResponse(c, "Prompt is too short")
		return
	}

	// 保存到文件
	promptPath := "config/variant_prompt.txt"
	if err := os.WriteFile(promptPath, []byte(req.Prompt), 0644); err != nil {
		logrus.Errorf("Failed to write prompt file: %v", err)
		InternalErrorResponse(c, "Failed to save prompt")
		return
	}

	// 更新内存中的配置
	config.VariantPromptTemplate = req.Prompt

	logrus.Infof("Variant prompt template updated successfully")

	response := map[string]interface{}{
		"message":    "Prompt updated successfully",
		"length":     len(req.Prompt),
		"updated_at": "now",
	}

	SuccessResponse(c, response)
}
