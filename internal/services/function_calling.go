package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"search-ec2/internal/config"
	"search-ec2/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

// FunctionCallingService Function Calling 解析服务
type FunctionCallingService struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

// NewFunctionCallingService 创建 Function Calling 服务
func NewFunctionCallingService() *FunctionCallingService {
	return &FunctionCallingService{
		client: &http.Client{
			Timeout: time.Duration(config.AppConfig.OpenAI.Timeout) * time.Second,
		},
		baseURL: config.AppConfig.OpenAI.BaseURL,
		apiKey:  config.AppConfig.OpenAI.APIKey,
		model:   config.AppConfig.OpenAI.ChatModel,
	}
}

// ParseQuery 解析用户查询意图
func (s *FunctionCallingService) ParseQuery(query string) (*models.ParsedQuery, error) {
	// 构建系统提示
	systemPrompt := `你是一个专业的商品搜索查询解析助手。你需要分析用户的自然语言查询，提取出商品类型、属性和过滤条件。

请仔细分析用户的查询意图，准确提取以下信息：
- 商品类型：用户想要搜索的商品类别
- 颜色、品牌、尺寸、材质、风格等属性
- 价格范围、使用场合、性别等过滤条件

如果某些信息在查询中没有明确提及，请不要添加或猜测。`

	// 构建用户消息
	userMessage := fmt.Sprintf("请解析这个商品搜索查询：%s", query)

	// 构建 Function 定义
	function := models.OpenAIFunction{
		Name:        config.FunctionSchema.FunctionName,
		Description: config.FunctionSchema.Description,
		Parameters:  config.FunctionSchema.Parameters,
	}

	// 构建请求
	request := models.OpenAIRequest{
		Model: s.model,
		Messages: []models.OpenAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Functions: []models.OpenAIFunction{function},
		FunctionCall: map[string]string{
			"name": config.FunctionSchema.FunctionName,
		},
		MaxTokens:   config.AppConfig.OpenAI.MaxTokens,
		Temperature: 0.1, // 低温度确保一致性
	}

	// 发送请求
	response, err := s.sendChatRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}

	// 解析 Function Call 结果
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	choice := response.Choices[0]
	if choice.Message.FunctionCall == nil {
		return nil, fmt.Errorf("no function call in response")
	}

	// 解析参数
	var parsedQuery models.ParsedQuery
	if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &parsedQuery); err != nil {
		return nil, fmt.Errorf("failed to parse function arguments: %w", err)
	}

	logrus.Debugf("Parsed query: %+v", parsedQuery)
	return &parsedQuery, nil
}

// sendChatRequest 发送聊天请求
func (s *FunctionCallingService) sendChatRequest(request models.OpenAIRequest) (*models.OpenAIResponse, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/chat/completions", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// 发送请求
	logrus.Debugf("Sending function calling request")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("OpenAI API error: %s", string(responseBody))
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	// 解析响应
	var response models.OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查 API 错误
	if response.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	logrus.Debugf("Function calling completed, used %d tokens", response.Usage.TotalTokens)
	return &response, nil
}

// HealthCheck 健康检查
func (s *FunctionCallingService) HealthCheck() error {
	// 尝试解析一个简单查询
	_, err := s.ParseQuery("测试查询")
	return err
}

// ValidateQuery 验证查询解析结果
func (s *FunctionCallingService) ValidateQuery(parsedQuery *models.ParsedQuery) error {
	// 检查必需字段
	if parsedQuery.ProductType == "" {
		return fmt.Errorf("product_type is required")
	}

	// 验证价格范围
	if parsedQuery.PriceMin != nil && parsedQuery.PriceMax != nil {
		if *parsedQuery.PriceMin > *parsedQuery.PriceMax {
			return fmt.Errorf("price_min cannot be greater than price_max")
		}
	}

	// 验证价格值
	if parsedQuery.PriceMin != nil && *parsedQuery.PriceMin < 0 {
		return fmt.Errorf("price_min cannot be negative")
	}

	if parsedQuery.PriceMax != nil && *parsedQuery.PriceMax < 0 {
		return fmt.Errorf("price_max cannot be negative")
	}

	return nil
}

// EnhanceQuery 增强查询解析（添加同义词、纠错等）
func (s *FunctionCallingService) EnhanceQuery(parsedQuery *models.ParsedQuery) *models.ParsedQuery {
	enhanced := *parsedQuery

	// 商品类型同义词映射
	productTypeMap := map[string]string{
		"牛仔服":  "牛仔裤",
		"丹宁裤":  "牛仔裤",
		"T恤衫":  "T恤",
		"短袖":   "T恤",
		"运动鞋":  "鞋子",
		"跑步鞋":  "鞋子",
		"球鞋":   "鞋子",
		"手机":   "智能手机",
		"电话":   "智能手机",
	}

	if synonym, exists := productTypeMap[enhanced.ProductType]; exists {
		enhanced.ProductType = synonym
	}

	// 颜色标准化
	colorMap := map[string]string{
		"红":    "红色",
		"蓝":    "蓝色",
		"黑":    "黑色",
		"白":    "白色",
		"绿":    "绿色",
		"黄":    "黄色",
		"紫":    "紫色",
		"粉":    "粉色",
		"灰":    "灰色",
		"棕":    "棕色",
		"深蓝":   "深蓝色",
		"浅蓝":   "浅蓝色",
		"天蓝":   "天蓝色",
		"海蓝":   "海蓝色",
	}

	if standardColor, exists := colorMap[enhanced.Color]; exists {
		enhanced.Color = standardColor
	}

	// 尺寸标准化
	sizeMap := map[string]string{
		"小":    "S",
		"中":    "M",
		"大":    "L",
		"特大":   "XL",
		"超大":   "XXL",
		"小号":   "S",
		"中号":   "M",
		"大号":   "L",
		"特大号":  "XL",
	}

	if standardSize, exists := sizeMap[enhanced.Size]; exists {
		enhanced.Size = standardSize
	}

	return &enhanced
}

// GetQuerySuggestions 获取查询建议
func (s *FunctionCallingService) GetQuerySuggestions(query string) ([]string, error) {
	systemPrompt := `你是一个商品搜索建议助手。基于用户的部分查询，生成5个相关的完整搜索建议。

要求：
1. 建议应该是完整的、自然的中文查询
2. 涵盖不同的商品属性和价格范围
3. 建议应该实用且常见
4. 每个建议不超过20个字

请以JSON数组格式返回建议列表。`

	userMessage := fmt.Sprintf("基于这个查询片段生成搜索建议：%s", query)

	request := models.OpenAIRequest{
		Model: s.model,
		Messages: []models.OpenAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	}

	response, err := s.sendChatRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no suggestions returned")
	}

	content := response.Choices[0].Message.Content

	// 尝试解析 JSON 数组
	var suggestions []string
	if err := json.Unmarshal([]byte(content), &suggestions); err != nil {
		// 如果解析失败，返回基础建议
		logrus.Warnf("Failed to parse suggestions JSON: %v", err)
		return s.getDefaultSuggestions(query), nil
	}

	return suggestions, nil
}

// getDefaultSuggestions 获取默认建议
func (s *FunctionCallingService) getDefaultSuggestions(query string) []string {
	return []string{
		query + " 黑色",
		query + " 白色",
		query + " 100元以下",
		query + " 品牌",
		query + " 大码",
	}
}
