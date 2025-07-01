package models

// SearchRequest 搜索请求
type SearchRequest struct {
	Query  string `json:"query" binding:"required"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query     string          `json:"query"`
	Total     int             `json:"total"`
	Results   []SearchResult  `json:"results"`
	ParsedQuery *ParsedQuery  `json:"parsed_query,omitempty"`
	TimeTaken int64           `json:"time_taken_ms"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Product     *Product `json:"product"`
	Score       float64  `json:"score"`
	MatchReason string   `json:"match_reason"`
	Variant     string   `json:"variant,omitempty"` // 匹配的变体文本
}

// ParsedQuery Function Calling 解析结果
type ParsedQuery struct {
	ProductType string                 `json:"product_type,omitempty"`
	Color       string                 `json:"color,omitempty"`
	PriceMin    *float64               `json:"price_min,omitempty"`
	PriceMax    *float64               `json:"price_max,omitempty"`
	Brand       string                 `json:"brand,omitempty"`
	Size        string                 `json:"size,omitempty"`
	Material    string                 `json:"material,omitempty"`
	Style       string                 `json:"style,omitempty"`
	Occasion    string                 `json:"occasion,omitempty"`
	Gender      string                 `json:"gender,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"` // 其他动态过滤条件
}

// SearchSuggestionsRequest 搜索建议请求
type SearchSuggestionsRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit,omitempty"`
}

// SearchSuggestionsResponse 搜索建议响应
type SearchSuggestionsResponse struct {
	Query       string   `json:"query"`
	Suggestions []string `json:"suggestions"`
}

// ToQdrantFilter 转换为 Qdrant 过滤条件
func (pq *ParsedQuery) ToQdrantFilter() map[string]interface{} {
	filter := make(map[string]interface{})
	must := []map[string]interface{}{}

	// 添加基础字段过滤
	if pq.Color != "" {
		must = append(must, map[string]interface{}{
			"key":   "color",
			"match": map[string]interface{}{"value": pq.Color},
		})
	}

	if pq.Brand != "" {
		must = append(must, map[string]interface{}{
			"key":   "brand",
			"match": map[string]interface{}{"value": pq.Brand},
		})
	}

	if pq.Size != "" {
		must = append(must, map[string]interface{}{
			"key":   "size",
			"match": map[string]interface{}{"value": pq.Size},
		})
	}

	if pq.Material != "" {
		must = append(must, map[string]interface{}{
			"key":   "material",
			"match": map[string]interface{}{"value": pq.Material},
		})
	}

	if pq.Style != "" {
		must = append(must, map[string]interface{}{
			"key":   "style",
			"match": map[string]interface{}{"value": pq.Style},
		})
	}

	if pq.Occasion != "" {
		must = append(must, map[string]interface{}{
			"key":   "occasion",
			"match": map[string]interface{}{"value": pq.Occasion},
		})
	}

	if pq.Gender != "" {
		must = append(must, map[string]interface{}{
			"key":   "gender",
			"match": map[string]interface{}{"value": pq.Gender},
		})
	}

	// 价格范围过滤
	if pq.PriceMin != nil || pq.PriceMax != nil {
		priceRange := map[string]interface{}{"key": "price"}
		rangeCondition := map[string]interface{}{}

		if pq.PriceMin != nil {
			rangeCondition["gte"] = *pq.PriceMin
		}
		if pq.PriceMax != nil {
			rangeCondition["lte"] = *pq.PriceMax
		}

		priceRange["range"] = rangeCondition
		must = append(must, priceRange)
	}

	// 状态过滤 - 只返回活跃商品
	must = append(must, map[string]interface{}{
		"key":   "status",
		"match": map[string]interface{}{"value": "active"},
	})

	// 添加动态过滤条件
	for key, value := range pq.Filters {
		must = append(must, map[string]interface{}{
			"key":   key,
			"match": map[string]interface{}{"value": value},
		})
	}

	if len(must) > 0 {
		filter["must"] = must
	}

	return filter
}

// GetSearchQuery 获取用于向量检索的查询文本
func (pq *ParsedQuery) GetSearchQuery() string {
	if pq.ProductType != "" {
		return pq.ProductType
	}
	return ""
}
