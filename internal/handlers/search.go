package handlers

import (
	"fmt"
	"search-ec2/internal/config"
	"search-ec2/internal/models"
	"search-ec2/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	serviceManager *services.ServiceManager
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(serviceManager *services.ServiceManager) *SearchHandler {
	return &SearchHandler{
		serviceManager: serviceManager,
	}
}

// Search 自然语言搜索
func (h *SearchHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.Query == "" {
		BadRequestResponse(c, "Query is required")
		return
	}

	if req.Limit <= 0 {
		req.Limit = config.AppConfig.Search.MaxResults
	}
	if req.Limit > config.AppConfig.Search.MaxResults {
		req.Limit = config.AppConfig.Search.MaxResults
	}

	startTime := time.Now()
	logrus.Infof("Processing search query: %s", req.Query)

	// 1. 解析用户查询意图
	parsedQuery, err := h.serviceManager.FunctionCalling.ParseQuery(req.Query)
	if err != nil {
		logrus.Errorf("Failed to parse query: %v", err)
		// 如果解析失败，使用原始查询进行向量搜索
		parsedQuery = &models.ParsedQuery{
			ProductType: req.Query,
		}
	}

	// 2. 验证解析结果
	if err := h.serviceManager.FunctionCalling.ValidateQuery(parsedQuery); err != nil {
		logrus.Warnf("Query validation failed: %v", err)
		// 继续处理，但记录警告
	}

	// 3. 增强查询（同义词、纠错等）
	enhancedQuery := h.serviceManager.FunctionCalling.EnhanceQuery(parsedQuery)

	// 4. 生成搜索向量
	searchText := enhancedQuery.GetSearchQuery()
	if searchText == "" {
		searchText = req.Query
	}

	queryVector, err := h.serviceManager.Embedding.GetEmbedding(searchText)
	if err != nil {
		logrus.Errorf("Failed to generate query vector: %v", err)
		InternalErrorResponse(c, "Failed to process search query")
		return
	}

	// 5. 构建过滤条件
	filter := enhancedQuery.ToQdrantFilter()

	// 6. 执行混合检索
	results, err := h.serviceManager.Qdrant.SearchProducts(queryVector, filter, req.Limit)
	if err != nil {
		logrus.Errorf("Failed to search products: %v", err)
		InternalErrorResponse(c, "Search failed")
		return
	}

	// 7. 构建响应
	timeTaken := time.Since(startTime).Milliseconds()
	
	response := models.SearchResponse{
		Query:       req.Query,
		Total:       len(results),
		Results:     results,
		ParsedQuery: enhancedQuery,
		TimeTaken:   timeTaken,
	}

	logrus.Infof("Search completed: query='%s', results=%d, time=%dms", 
		req.Query, len(results), timeTaken)

	SuccessResponse(c, response)
}

// GetSuggestions 获取搜索建议
func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		BadRequestResponse(c, "Query parameter is required")
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit := 5
	if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	logrus.Infof("Generating suggestions for query: %s", query)

	// 使用 Function Calling 服务生成建议
	suggestions, err := h.serviceManager.FunctionCalling.GetQuerySuggestions(query)
	if err != nil {
		logrus.Errorf("Failed to generate suggestions: %v", err)
		// 返回默认建议
		suggestions = h.getDefaultSuggestions(query, limit)
	}

	// 限制建议数量
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	response := models.SearchSuggestionsResponse{
		Query:       query,
		Suggestions: suggestions,
	}

	SuccessResponse(c, response)
}

// getDefaultSuggestions 获取默认搜索建议
func (h *SearchHandler) getDefaultSuggestions(query string, limit int) []string {
	// 基于查询生成一些基础建议
	suggestions := []string{
		query + " 黑色",
		query + " 白色",
		query + " 蓝色",
		query + " 100元以下",
		query + " 200元以下",
		query + " 品牌",
		query + " 大码",
		query + " 小码",
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}
