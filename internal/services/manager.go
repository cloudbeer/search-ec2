package services

import (
	"fmt"
	"search-ec2/internal/config"

	"github.com/sirupsen/logrus"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	Qdrant            *QdrantService
	Embedding         *CachedEmbeddingService
	FunctionCalling   *FunctionCallingService
	VariantGeneration *VariantGenerationService
}

// NewServiceManager 创建服务管理器
func NewServiceManager() (*ServiceManager, error) {
	logrus.Info("Initializing services...")

	// 初始化 Qdrant 服务
	qdrantService, err := NewQdrantService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Qdrant service: %w", err)
	}
	logrus.Info("Qdrant service initialized")

	// 初始化向量化服务
	embeddingService := NewCachedEmbeddingService()
	logrus.Info("Embedding service initialized")

	// 初始化 Function Calling 服务
	functionCallingService := NewFunctionCallingService()
	logrus.Info("Function calling service initialized")

	// 初始化变体生成服务
	variantGenerationService := NewVariantGenerationService()
	logrus.Info("Variant generation service initialized")

	manager := &ServiceManager{
		Qdrant:            qdrantService,
		Embedding:         embeddingService,
		FunctionCalling:   functionCallingService,
		VariantGeneration: variantGenerationService,
	}

	logrus.Info("All services initialized successfully")
	return manager, nil
}

// HealthCheck 检查所有服务健康状态
func (sm *ServiceManager) HealthCheck() map[string]string {
	status := make(map[string]string)

	// 检查 Qdrant
	if err := sm.Qdrant.HealthCheck(); err != nil {
		status["qdrant"] = fmt.Sprintf("error: %v", err)
	} else {
		status["qdrant"] = "ok"
	}

	// 检查向量化服务（如果配置了 OpenAI API Key）
	if config.AppConfig.OpenAI.APIKey != "" {
		if err := sm.Embedding.HealthCheck(); err != nil {
			status["embedding"] = fmt.Sprintf("error: %v", err)
		} else {
			status["embedding"] = "ok"
		}

		// 检查 Function Calling 服务
		if err := sm.FunctionCalling.HealthCheck(); err != nil {
			status["function_calling"] = fmt.Sprintf("error: %v", err)
		} else {
			status["function_calling"] = "ok"
		}

		// 检查变体生成服务
		if err := sm.VariantGeneration.HealthCheck(); err != nil {
			status["variant_generation"] = fmt.Sprintf("error: %v", err)
		} else {
			status["variant_generation"] = "ok"
		}
	} else {
		status["embedding"] = "not_configured"
		status["function_calling"] = "not_configured"
		status["variant_generation"] = "not_configured"
	}

	return status
}

// GetStats 获取系统统计信息
func (sm *ServiceManager) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Qdrant 统计
	qdrantStats, err := sm.Qdrant.GetStats()
	if err != nil {
		logrus.Errorf("Failed to get Qdrant stats: %v", err)
		stats["qdrant"] = map[string]interface{}{"error": err.Error()}
	} else {
		stats["qdrant"] = qdrantStats
	}

	// 向量化缓存统计
	stats["embedding_cache"] = map[string]interface{}{
		"size": sm.Embedding.cache.Size(),
	}

	// 配置信息
	stats["config"] = map[string]interface{}{
		"vector_size":          config.AppConfig.Qdrant.VectorSize,
		"collection_name":      config.AppConfig.Qdrant.CollectionName,
		"embedding_model":      config.AppConfig.OpenAI.EmbeddingModel,
		"chat_model":          config.AppConfig.OpenAI.ChatModel,
		"max_results":         config.AppConfig.Search.MaxResults,
		"similarity_threshold": config.AppConfig.Search.SimilarityThreshold,
	}

	// 功能开关状态
	stats["features"] = map[string]interface{}{
		"batch_import":       config.AppConfig.Features.EnableBatchImport,
		"variant_generation": config.AppConfig.Features.EnableVariantGeneration,
		"function_calling":   config.AppConfig.Features.EnableFunctionCalling,
		"search_suggestions": config.AppConfig.Features.EnableSearchSuggestions,
	}

	return stats, nil
}

// Close 关闭所有服务连接
func (sm *ServiceManager) Close() error {
	logrus.Info("Closing services...")
	
	// 清理缓存
	sm.Embedding.cache.Clear()
	
	logrus.Info("All services closed")
	return nil
}
