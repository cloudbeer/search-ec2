package handlers

import (
	"search-ec2/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由（向后兼容）
func SetupRoutes(r *gin.Engine) {
	SetupRoutesWithServices(r, nil)
}

// SetupRoutesWithServices 设置路由（带服务管理器）
func SetupRoutesWithServices(r *gin.Engine, serviceManager *services.ServiceManager) {
	// 创建处理器实例
	healthHandler := NewHealthHandler(serviceManager)
	
	var productHandler *ProductHandler
	var searchHandler *SearchHandler
	var configHandler *ConfigHandler
	
	if serviceManager != nil {
		productHandler = NewProductHandler(serviceManager)
		searchHandler = NewSearchHandler(serviceManager)
		configHandler = NewConfigHandler(serviceManager)
	}

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", healthHandler.Check)

		// 商品管理路由组
		products := api.Group("/products")
		{
			if productHandler != nil {
				products.POST("", productHandler.CreateProduct)
				products.GET("/:id", productHandler.GetProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
				products.POST("/batch", productHandler.BatchImport)
				products.POST("/:id/variants/regenerate", productHandler.RegenerateVariants)
			} else {
				// 备用 TODO 响应
				products.POST("", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Create product - TODO"})
				})
				products.GET("/:id", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Get product - TODO"})
				})
				products.PUT("/:id", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Update product - TODO"})
				})
				products.DELETE("/:id", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Delete product - TODO"})
				})
				products.POST("/batch", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Batch import - TODO"})
				})
				products.POST("/:id/variants/regenerate", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Regenerate variants - TODO"})
				})
			}
		}

		// 搜索路由组
		search := api.Group("/search")
		{
			if searchHandler != nil {
				search.POST("", searchHandler.Search)
				search.GET("/suggestions", searchHandler.GetSuggestions)
			} else {
				// 备用 TODO 响应
				search.POST("", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Search products - TODO"})
				})
				search.GET("/suggestions", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Search suggestions - TODO"})
				})
			}
		}

		// 配置管理路由组
		config := api.Group("/config")
		{
			if configHandler != nil {
				config.GET("/function-schema", configHandler.GetFunctionSchema)
				config.PUT("/function-schema", configHandler.UpdateFunctionSchema)
				config.GET("/variant-prompt", configHandler.GetVariantPrompt)
				config.PUT("/variant-prompt", configHandler.UpdateVariantPrompt)
			} else {
				// 备用 TODO 响应
				config.GET("/function-schema", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Get function schema - TODO"})
				})
				config.PUT("/function-schema", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Update function schema - TODO"})
				})
				config.GET("/variant-prompt", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Get variant prompt - TODO"})
				})
				config.PUT("/variant-prompt", func(c *gin.Context) {
					SuccessResponse(c, gin.H{"message": "Update variant prompt - TODO"})
				})
			}
		}

		// 系统统计
		if serviceManager != nil {
			api.GET("/stats", func(c *gin.Context) {
				stats, err := serviceManager.GetStats()
				if err != nil {
					InternalErrorResponse(c, "Failed to get stats")
					return
				}
				SuccessResponse(c, stats)
			})
		} else {
			api.GET("/stats", func(c *gin.Context) {
				SuccessResponse(c, gin.H{"message": "System stats - TODO"})
			})
		}
	}

	// 根路径
	r.GET("/", func(c *gin.Context) {
		SuccessResponse(c, gin.H{
			"name":    "Search EC2 - Natural Language Product Search System",
			"version": "1.0.0",
			"status":  "running",
		})
	})
}
