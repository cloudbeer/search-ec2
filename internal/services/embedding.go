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

// EmbeddingService 向量化服务
type EmbeddingService struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

// NewEmbeddingService 创建向量化服务
func NewEmbeddingService() *EmbeddingService {
	return &EmbeddingService{
		client: &http.Client{
			Timeout: time.Duration(config.AppConfig.OpenAI.Timeout) * time.Second,
		},
		baseURL: config.AppConfig.OpenAI.BaseURL,
		apiKey:  config.AppConfig.OpenAI.APIKey,
		model:   config.AppConfig.OpenAI.EmbeddingModel,
	}
}

// GetEmbedding 获取单个文本的向量
func (s *EmbeddingService) GetEmbedding(text string) ([]float32, error) {
	embeddings, err := s.GetEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embeddings[0], nil
}

// GetEmbeddings 批量获取文本向量
func (s *EmbeddingService) GetEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// 构建请求
	request := models.EmbeddingRequest{
		Model: s.model,
		Input: texts,
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/embeddings", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// 发送请求
	logrus.Debugf("Sending embedding request for %d texts", len(texts))
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
	var response models.EmbeddingResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查 API 错误
	if response.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	// 提取向量
	embeddings := make([][]float32, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	logrus.Debugf("Successfully got embeddings for %d texts, used %d tokens", 
		len(embeddings), response.Usage.TotalTokens)

	return embeddings, nil
}

// GetProductVariantEmbeddings 为商品变体生成向量
func (s *EmbeddingService) GetProductVariantEmbeddings(variants []string) ([]models.ProductVariant, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants provided")
	}

	// 获取向量
	embeddings, err := s.GetEmbeddings(variants)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	// 创建变体对象
	productVariants := make([]models.ProductVariant, len(variants))
	for i, variant := range variants {
		productVariants[i] = models.ProductVariant{
			ID:          fmt.Sprintf("variant_%d_%d", time.Now().UnixNano(), i),
			Text:        variant,
			Vector:      embeddings[i],
			GeneratedAt: time.Now(),
		}
	}

	return productVariants, nil
}

// HealthCheck 健康检查
func (s *EmbeddingService) HealthCheck() error {
	// 尝试获取一个简单文本的向量
	_, err := s.GetEmbedding("test")
	return err
}

// BatchEmbedding 批量向量化处理（支持大量文本）
func (s *EmbeddingService) BatchEmbedding(texts []string, batchSize int) ([][]float32, error) {
	if batchSize <= 0 {
		batchSize = 100 // 默认批次大小
	}

	var allEmbeddings [][]float32

	// 分批处理
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := s.GetEmbeddings(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)

		// 添加延迟以避免速率限制
		if i+batchSize < len(texts) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return allEmbeddings, nil
}

// EmbeddingCache 向量缓存（简单内存缓存）
type EmbeddingCache struct {
	cache map[string][]float32
}

// NewEmbeddingCache 创建向量缓存
func NewEmbeddingCache() *EmbeddingCache {
	return &EmbeddingCache{
		cache: make(map[string][]float32),
	}
}

// Get 获取缓存的向量
func (c *EmbeddingCache) Get(text string) ([]float32, bool) {
	embedding, exists := c.cache[text]
	return embedding, exists
}

// Set 设置缓存的向量
func (c *EmbeddingCache) Set(text string, embedding []float32) {
	c.cache[text] = embedding
}

// Clear 清空缓存
func (c *EmbeddingCache) Clear() {
	c.cache = make(map[string][]float32)
}

// Size 获取缓存大小
func (c *EmbeddingCache) Size() int {
	return len(c.cache)
}

// CachedEmbeddingService 带缓存的向量化服务
type CachedEmbeddingService struct {
	*EmbeddingService
	cache *EmbeddingCache
}

// NewCachedEmbeddingService 创建带缓存的向量化服务
func NewCachedEmbeddingService() *CachedEmbeddingService {
	return &CachedEmbeddingService{
		EmbeddingService: NewEmbeddingService(),
		cache:           NewEmbeddingCache(),
	}
}

// GetEmbedding 获取向量（带缓存）
func (s *CachedEmbeddingService) GetEmbedding(text string) ([]float32, error) {
	// 检查缓存
	if embedding, exists := s.cache.Get(text); exists {
		logrus.Debugf("Cache hit for text: %s", text[:min(50, len(text))])
		return embedding, nil
	}

	// 获取向量
	embedding, err := s.EmbeddingService.GetEmbedding(text)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	s.cache.Set(text, embedding)
	return embedding, nil
}

// GetEmbeddings 批量获取向量（带缓存）
func (s *CachedEmbeddingService) GetEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	uncachedTexts := []string{}
	uncachedIndices := []int{}

	// 检查缓存
	for i, text := range texts {
		if embedding, exists := s.cache.Get(text); exists {
			embeddings[i] = embedding
		} else {
			uncachedTexts = append(uncachedTexts, text)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	// 获取未缓存的向量
	if len(uncachedTexts) > 0 {
		newEmbeddings, err := s.EmbeddingService.GetEmbeddings(uncachedTexts)
		if err != nil {
			return nil, err
		}

		// 填充结果并缓存
		for i, embedding := range newEmbeddings {
			index := uncachedIndices[i]
			text := uncachedTexts[i]
			embeddings[index] = embedding
			s.cache.Set(text, embedding)
		}
	}

	logrus.Debugf("Processed %d texts, %d from cache, %d new", 
		len(texts), len(texts)-len(uncachedTexts), len(uncachedTexts))

	return embeddings, nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
