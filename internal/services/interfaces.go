package services

import "search-ec2/internal/models"

// EmbeddingServiceInterface 向量化服务接口
type EmbeddingServiceInterface interface {
	GetEmbedding(text string) ([]float32, error)
	GetEmbeddings(texts []string) ([][]float32, error)
	GetProductVariantEmbeddings(variants []string) ([]models.ProductVariant, error)
	HealthCheck() error
}
