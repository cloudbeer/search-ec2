package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"search-ec2/internal/config"
	"search-ec2/internal/models"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// VariantGenerationService AI 变体生成服务
type VariantGenerationService struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

// NewVariantGenerationService 创建变体生成服务
func NewVariantGenerationService() *VariantGenerationService {
	return &VariantGenerationService{
		client: &http.Client{
			Timeout: time.Duration(config.AppConfig.OpenAI.Timeout) * time.Second,
		},
		baseURL: config.AppConfig.OpenAI.BaseURL,
		apiKey:  config.AppConfig.OpenAI.APIKey,
		model:   config.AppConfig.OpenAI.ChatModel,
	}
}

// GenerateVariants 为商品生成变体描述
func (s *VariantGenerationService) GenerateVariants(product *models.Product, variantCount int) ([]string, error) {
	if variantCount <= 0 {
		variantCount = 5 // 默认生成5个变体
	}

	if variantCount > 20 {
		variantCount = 20 // 最多20个变体
	}

	// 构建提示词
	prompt := s.buildPrompt(product, variantCount)

	// 构建请求
	request := models.OpenAIRequest{
		Model: s.model,
		Messages: []models.OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1500,
		Temperature: 0.8, // 较高温度增加创造性
	}

	// 发送请求
	response, err := s.sendChatRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate variants: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	content := response.Choices[0].Message.Content

	// 解析变体
	variants, err := s.parseVariants(content)
	if err != nil {
		logrus.Warnf("Failed to parse variants JSON, using fallback: %v", err)
		// 如果解析失败，使用备用方法
		variants = s.parseVariantsFallback(content)
	}

	// 过滤和验证变体
	validVariants := s.filterVariants(variants, product)

	logrus.Infof("Generated %d variants for product %s", len(validVariants), product.ID)
	return validVariants, nil
}

// buildPrompt 构建变体生成提示词
func (s *VariantGenerationService) buildPrompt(product *models.Product, variantCount int) string {
	// 使用配置的提示词模板
	prompt := config.VariantPromptTemplate

	// 替换占位符
	replacements := map[string]string{
		"{variant_count}": fmt.Sprintf("%d", variantCount),
		"{product_name}":  product.Name,
		"{category}":      product.Category,
		"{color}":         product.Color,
		"{price}":         fmt.Sprintf("%.2f%s", product.Price, product.Currency),
		"{brand}":         product.Brand,
		"{size}":          product.Size,
		"{material}":      product.Material,
		"{description}":   product.Description,
	}

	for placeholder, value := range replacements {
		if value == "" {
			value = "未指定"
		}
		prompt = strings.ReplaceAll(prompt, placeholder, value)
	}

	return prompt
}

// parseVariants 解析变体 JSON 响应
func (s *VariantGenerationService) parseVariants(content string) ([]string, error) {
	// 尝试直接解析 JSON 数组
	var variants []string
	if err := json.Unmarshal([]byte(content), &variants); err == nil {
		return variants, nil
	}

	// 尝试从文本中提取 JSON 数组
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	
	if start != -1 && end != -1 && end > start {
		jsonStr := content[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &variants); err == nil {
			return variants, nil
		}
	}

	return nil, fmt.Errorf("failed to parse JSON variants")
}

// parseVariantsFallback 备用变体解析方法
func (s *VariantGenerationService) parseVariantsFallback(content string) []string {
	lines := strings.Split(content, "\n")
	var variants []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 跳过空行和非变体行
		if line == "" || strings.HasPrefix(line, "```") || 
		   strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// 移除序号、引号、逗号等
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimPrefix(line, "•")
		
		// 移除数字序号
		if len(line) > 2 && line[1] == '.' {
			line = line[2:]
		}
		
		line = strings.Trim(line, " \t\"',")
		
		if len(line) > 5 && len(line) < 100 { // 合理的变体长度
			variants = append(variants, line)
		}
	}

	return variants
}

// filterVariants 过滤和验证变体
func (s *VariantGenerationService) filterVariants(variants []string, product *models.Product) []string {
	var validVariants []string
	seen := make(map[string]bool)

	for _, variant := range variants {
		variant = strings.TrimSpace(variant)
		
		// 基本验证
		if len(variant) < 5 || len(variant) > 100 {
			continue
		}

		// 去重
		if seen[variant] {
			continue
		}
		seen[variant] = true

		// 检查是否包含核心商品信息
		if s.isValidVariant(variant, product) {
			validVariants = append(validVariants, variant)
		}
	}

	return validVariants
}

// isValidVariant 验证变体是否有效
func (s *VariantGenerationService) isValidVariant(variant string, product *models.Product) bool {
	variant = strings.ToLower(variant)
	
	// 检查是否包含商品名称或类别的关键词
	productName := strings.ToLower(product.Name)
	category := strings.ToLower(product.Category)
	
	// 提取关键词
	keywords := []string{}
	if productName != "" {
		keywords = append(keywords, strings.Fields(productName)...)
	}
	if category != "" {
		keywords = append(keywords, strings.Fields(category)...)
	}

	// 至少包含一个关键词
	for _, keyword := range keywords {
		if len(keyword) > 1 && strings.Contains(variant, keyword) {
			return true
		}
	}

	return false
}

// GenerateVariantsWithEmbeddings 生成变体并获取向量
func (s *VariantGenerationService) GenerateVariantsWithEmbeddings(
	product *models.Product, 
	variantCount int, 
	embeddingService EmbeddingServiceInterface,
) ([]models.ProductVariant, error) {
	
	// 生成变体文本
	variantTexts, err := s.GenerateVariants(product, variantCount)
	if err != nil {
		return nil, fmt.Errorf("failed to generate variant texts: %w", err)
	}

	if len(variantTexts) == 0 {
		return nil, fmt.Errorf("no valid variants generated")
	}

	// 获取向量
	productVariants, err := embeddingService.GetProductVariantEmbeddings(variantTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant embeddings: %w", err)
	}

	// 设置商品 ID
	for i := range productVariants {
		productVariants[i].ProductID = product.ID
	}

	return productVariants, nil
}

// RegenerateVariants 重新生成商品变体
func (s *VariantGenerationService) RegenerateVariants(
	productID string,
	variantCount int,
	qdrantService *QdrantService,
	embeddingService EmbeddingServiceInterface,
) error {
	
	// 获取商品信息
	product, err := qdrantService.GetProduct(productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// 删除旧变体
	if err := qdrantService.DeleteProduct(productID); err != nil {
		return fmt.Errorf("failed to delete old variants: %w", err)
	}

	// 生成新变体
	variants, err := s.GenerateVariantsWithEmbeddings(product, variantCount, embeddingService)
	if err != nil {
		return fmt.Errorf("failed to generate new variants: %w", err)
	}

	// 插入新变体
	if err := qdrantService.InsertProduct(product, variants); err != nil {
		return fmt.Errorf("failed to insert new variants: %w", err)
	}

	logrus.Infof("Regenerated %d variants for product %s", len(variants), productID)
	return nil
}

// sendChatRequest 发送聊天请求
func (s *VariantGenerationService) sendChatRequest(request models.OpenAIRequest) (*models.OpenAIResponse, error) {
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
	logrus.Debugf("Sending variant generation request")
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

	return &response, nil
}

// HealthCheck 健康检查
func (s *VariantGenerationService) HealthCheck() error {
	// 创建测试商品
	testProduct := &models.Product{
		ID:       "test",
		Name:     "测试商品",
		Category: "测试类别",
		Price:    100.0,
		Currency: "元",
	}

	// 尝试生成变体
	_, err := s.GenerateVariants(testProduct, 2)
	return err
}
