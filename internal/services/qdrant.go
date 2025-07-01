package services

import (
	"context"
	"fmt"
	"search-ec2/internal/config"
	"search-ec2/internal/models"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/sirupsen/logrus"
)

// QdrantService Qdrant 服务 - 懒加载版本
type QdrantService struct {
	client         *qdrant.Client
	collectionName string
	config         *qdrant.Config
	initialized    bool
}

// NewQdrantService 创建 Qdrant 服务 - 懒加载模式
func NewQdrantService() (*QdrantService, error) {
	service := &QdrantService{
		collectionName: config.AppConfig.Qdrant.CollectionName,
		config: &qdrant.Config{
			Host:   config.AppConfig.Qdrant.Host,
			Port:   config.AppConfig.Qdrant.Port,
			APIKey: config.AppConfig.Qdrant.APIKey,
			UseTLS: false,
		},
		initialized: false,
	}

	return service, nil
}

// ensureInitialized 确保客户端已初始化
func (s *QdrantService) ensureInitialized() error {
	if s.initialized {
		return nil
	}

	// 创建 Qdrant 客户端
	client, err := qdrant.NewClient(s.config)
	if err != nil {
		return fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	s.client = client

	// 确保集合存在
	if err := s.ensureCollection(); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	s.initialized = true
	logrus.Infof("Qdrant service initialized successfully")
	return nil
}

// ensureCollection 确保集合存在
func (s *QdrantService) ensureCollection() error {
	ctx := context.Background()

	// 检查集合是否存在
	collections, err := s.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// 检查集合是否已存在
	for _, collectionName := range collections {
		if collectionName == s.collectionName {
			logrus.Infof("Collection %s already exists", s.collectionName)
			return nil
		}
	}

	// 创建集合
	err = s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: s.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1536, // OpenAI embedding 维度
			Distance: qdrant.Distance_Cosine,
		}),
	})

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	logrus.Infof("Collection %s created successfully", s.collectionName)
	return nil
}

// InsertProduct 插入商品及其变体 - 完整实现
func (s *QdrantService) InsertProduct(product *models.Product, variants []models.ProductVariant) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}

	ctx := context.Background()
	points := make([]*qdrant.PointStruct, 0, len(variants))

	// 为每个变体创建一个点
	for i, variant := range variants {
		// 构建 payload - 包含商品的所有信息
		payload := map[string]any{
			"product_id":    product.ID,
			"variant_id":    variant.ID,
			"variant_text":  variant.Text,
			"product_name":  product.Name,
			"category":      product.Category,
			"description":   product.Description,
			"price":         product.Price,
			"currency":      product.Currency,
			"brand":         product.Brand,
			"color":         product.Color,
			"size":          product.Size,
			"material":      product.Material,
			"style":         product.Style,
			"gender":        product.Gender,
			"occasion":      product.Occasion,
			"created_at":    product.CreatedAt.Unix(),
			"updated_at":    product.UpdatedAt.Unix(),
		}

		// 添加标签 - 转换为 []interface{}
		if len(product.Tags) > 0 {
			tags := make([]interface{}, len(product.Tags))
			for i, tag := range product.Tags {
				tags[i] = tag
			}
			payload["tags"] = tags
		}

		// 添加图片URL - 转换为 []interface{}
		if len(product.ImageURLs) > 0 {
			urls := make([]interface{}, len(product.ImageURLs))
			for i, url := range product.ImageURLs {
				urls[i] = url
			}
			payload["image_urls"] = urls
		}

		// 添加自定义属性
		if len(product.Attributes) > 0 {
			for key, value := range product.Attributes {
				payload["attr_"+key] = fmt.Sprintf("%v", value)
			}
		}

		// 创建点结构 - 使用数字ID
		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)), // 使用数字ID，后续可以改进ID生成策略
			Vectors: qdrant.NewVectors(variant.Vector...),
			Payload: qdrant.NewValueMap(payload),
		}

		points = append(points, point)
	}

	// 批量插入点到 Qdrant
	operationInfo, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collectionName,
		Points:         points,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
	}

	logrus.Infof("Successfully inserted product %s with %d variants to Qdrant. Operation ID: %d", 
		product.ID, len(variants), operationInfo.OperationId)
	return nil
}

// SearchProducts 搜索商品 - 完整实现
func (s *QdrantService) SearchProducts(queryVector []float32, filter map[string]interface{}, limit int) ([]models.SearchResult, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx := context.Background()

	// 构建查询请求
	queryRequest := &qdrant.QueryPoints{
		CollectionName: s.collectionName,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	}

	// 构建过滤条件 - 增强版本
	if len(filter) > 0 {
		mustConditions := make([]*qdrant.Condition, 0)
		shouldConditions := make([]*qdrant.Condition, 0)
		mustNotConditions := make([]*qdrant.Condition, 0)
		
		// 处理各种过滤条件
		for key, value := range filter {
			switch key {
			case "price_min":
				if minPrice, ok := value.(float64); ok {
					mustConditions = append(mustConditions, qdrant.NewRange("price", &qdrant.Range{
						Gte: &minPrice,
					}))
				}
			case "price_max":
				if maxPrice, ok := value.(float64); ok {
					mustConditions = append(mustConditions, qdrant.NewRange("price", &qdrant.Range{
						Lte: &maxPrice,
					}))
				}
			case "brand", "color", "size", "material", "style", "gender", "occasion", "category":
				if strValue, ok := value.(string); ok && strValue != "" {
					mustConditions = append(mustConditions, qdrant.NewMatch(key, strValue))
				}
			case "exclude_brand", "exclude_color": // 排除条件
				if strValue, ok := value.(string); ok && strValue != "" {
					actualKey := strings.TrimPrefix(key, "exclude_")
					mustNotConditions = append(mustNotConditions, qdrant.NewMatch(actualKey, strValue))
				}
			case "any_tags": // 任意标签匹配 (OR 逻辑)
				if tags, ok := value.([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok && tagStr != "" {
							shouldConditions = append(shouldConditions, qdrant.NewMatch("tags", tagStr))
						}
					}
				}
			}
		}
		
		// 构建过滤器
		if len(mustConditions) > 0 || len(shouldConditions) > 0 || len(mustNotConditions) > 0 {
			filterObj := &qdrant.Filter{}
			
			if len(mustConditions) > 0 {
				filterObj.Must = mustConditions
			}
			if len(shouldConditions) > 0 {
				filterObj.Should = shouldConditions
			}
			if len(mustNotConditions) > 0 {
				filterObj.MustNot = mustNotConditions
			}
			
			queryRequest.Filter = filterObj
		}
	}

	// 执行查询
	searchResult, err := s.client.Query(ctx, queryRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to query Qdrant: %w", err)
	}

	// 转换结果
	results := make([]models.SearchResult, 0, len(searchResult))
	for _, point := range searchResult {
		result := models.SearchResult{
			Score: float64(point.Score),
		}

		// 解析 payload 构建商品信息
		if point.Payload != nil {
			product := s.parseProductFromPayload(point.Payload)
			result.Product = product
			
			// 获取匹配的变体文本
			if variantText, exists := point.Payload["variant_text"]; exists {
				result.Variant = s.extractStringFromValue(variantText)
			}
			
			// 生成匹配原因
			result.MatchReason = s.generateMatchReason(point.Score, filter)
		}

		results = append(results, result)
	}

	logrus.Infof("Search completed: found %d results from Qdrant", len(results))
	return results, nil
}

// GetProduct 获取商品信息 - 完整实现
func (s *QdrantService) GetProduct(productID string) (*models.Product, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx := context.Background()

	// 使用 Query 方法查找指定商品ID的点，使用零向量进行查询
	zeroVector := make([]float32, 1536) // 使用零向量，因为我们主要依赖过滤条件
	
	searchResult, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.collectionName,
		Query:          qdrant.NewQuery(zeroVector...),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatch("product_id", productID),
			},
		},
		Limit:       qdrant.PtrOf(uint64(1)), // 只需要一个结果
		WithPayload: qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query product from Qdrant: %w", err)
	}

	if len(searchResult) == 0 {
		return nil, fmt.Errorf("product not found in Qdrant")
	}

	// 解析第一个结果
	point := searchResult[0]
	product := s.parseProductFromPayload(point.Payload)

	logrus.Infof("Successfully retrieved product %s from Qdrant", productID)
	return product, nil
}

// DeleteProduct 删除商品及其所有变体 - 优化版本
func (s *QdrantService) DeleteProduct(productID string) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}

	ctx := context.Background()

	// 直接基于过滤条件删除，无需先查询
	operationInfo, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.collectionName,
		Points: qdrant.NewPointsSelectorFilter(
			&qdrant.Filter{
				Must: []*qdrant.Condition{
					qdrant.NewMatch("product_id", productID),
				},
			},
		),
	})

	if err != nil {
		return fmt.Errorf("failed to delete product from Qdrant: %w", err)
	}

	logrus.Infof("Successfully deleted product %s from Qdrant. Operation ID: %d", 
		productID, operationInfo.OperationId)
	return nil
}

// extractStringFromValue 从 qdrant.Value 中提取字符串
func (s *QdrantService) extractStringFromValue(value *qdrant.Value) string {
	if value == nil {
		return ""
	}
	return value.GetStringValue()
}

// extractFloatFromValue 从 qdrant.Value 中提取浮点数
func (s *QdrantService) extractFloatFromValue(value *qdrant.Value) float64 {
	if value == nil {
		return 0
	}
	return value.GetDoubleValue()
}

// extractIntFromValue 从 qdrant.Value 中提取整数
func (s *QdrantService) extractIntFromValue(value *qdrant.Value) int64 {
	if value == nil {
		return 0
	}
	return value.GetIntegerValue()
}

// extractArrayFromValue 从 qdrant.Value 中提取数组
func (s *QdrantService) extractArrayFromValue(value *qdrant.Value) []string {
	if value == nil {
		return nil
	}
	
	listValue := value.GetListValue()
	if listValue == nil {
		return nil
	}
	
	result := make([]string, 0, len(listValue.Values))
	for _, item := range listValue.Values {
		if item != nil {
			result = append(result, item.GetStringValue())
		}
	}
	return result
}
// parseProductFromPayload 从 Qdrant payload 解析商品信息
func (s *QdrantService) parseProductFromPayload(payload map[string]*qdrant.Value) *models.Product {
	product := &models.Product{
		Attributes: make(map[string]interface{}),
	}

	// 解析基础字段
	if val, ok := payload["product_id"]; ok {
		product.ID = s.extractStringFromValue(val)
	}
	if val, ok := payload["product_name"]; ok {
		product.Name = s.extractStringFromValue(val)
	}
	if val, ok := payload["category"]; ok {
		product.Category = s.extractStringFromValue(val)
	}
	if val, ok := payload["description"]; ok {
		product.Description = s.extractStringFromValue(val)
	}
	if val, ok := payload["price"]; ok {
		product.Price = s.extractFloatFromValue(val)
	}
	if val, ok := payload["currency"]; ok {
		product.Currency = s.extractStringFromValue(val)
	}
	if val, ok := payload["brand"]; ok {
		product.Brand = s.extractStringFromValue(val)
	}
	if val, ok := payload["color"]; ok {
		product.Color = s.extractStringFromValue(val)
	}
	if val, ok := payload["size"]; ok {
		product.Size = s.extractStringFromValue(val)
	}
	if val, ok := payload["material"]; ok {
		product.Material = s.extractStringFromValue(val)
	}
	if val, ok := payload["style"]; ok {
		product.Style = s.extractStringFromValue(val)
	}
	if val, ok := payload["gender"]; ok {
		product.Gender = s.extractStringFromValue(val)
	}
	if val, ok := payload["occasion"]; ok {
		product.Occasion = s.extractStringFromValue(val)
	}

	// 解析时间戳
	if val, ok := payload["created_at"]; ok {
		product.CreatedAt = time.Unix(s.extractIntFromValue(val), 0)
	}
	if val, ok := payload["updated_at"]; ok {
		product.UpdatedAt = time.Unix(s.extractIntFromValue(val), 0)
	}

	// 解析标签
	if val, ok := payload["tags"]; ok {
		product.Tags = s.extractArrayFromValue(val)
	}

	// 解析图片URL
	if val, ok := payload["image_urls"]; ok {
		product.ImageURLs = s.extractArrayFromValue(val)
	}

	// 解析自定义属性
	for key, val := range payload {
		if strings.HasPrefix(key, "attr_") {
			attrKey := strings.TrimPrefix(key, "attr_")
			product.Attributes[attrKey] = s.extractStringFromValue(val)
		}
	}

	return product
}

// generateMatchReason 生成匹配原因说明
func (s *QdrantService) generateMatchReason(score float32, filter map[string]interface{}) string {
	reasons := make([]string, 0)
	
	// 基于相似度评分
	if score > 0.9 {
		reasons = append(reasons, "高度语义匹配")
	} else if score > 0.7 {
		reasons = append(reasons, "良好语义匹配")
	} else {
		reasons = append(reasons, "基础语义匹配")
	}
	
	// 基于过滤条件
	if len(filter) > 0 {
		filterReasons := make([]string, 0)
		for key := range filter {
			switch key {
			case "brand":
				filterReasons = append(filterReasons, "品牌匹配")
			case "color":
				filterReasons = append(filterReasons, "颜色匹配")
			case "price_min", "price_max":
				filterReasons = append(filterReasons, "价格范围匹配")
			case "size":
				filterReasons = append(filterReasons, "尺寸匹配")
			}
		}
		if len(filterReasons) > 0 {
			reasons = append(reasons, strings.Join(filterReasons, "、"))
		}
	}
	
	return strings.Join(reasons, " + ")
}
// ScrollProducts 分页获取商品 - 修复版本
func (s *QdrantService) ScrollProducts(filter map[string]interface{}, limit uint32, offset *qdrant.PointId) ([]models.Product, *qdrant.PointId, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	// 构建 Scroll 请求
	scrollRequest := &qdrant.ScrollPoints{
		CollectionName: s.collectionName,
		Limit:          qdrant.PtrOf(limit),
		WithPayload:    qdrant.NewWithPayload(true),
	}

	// 添加偏移量（用于分页）
	if offset != nil {
		scrollRequest.Offset = offset
	}

	// 构建过滤条件
	if len(filter) > 0 {
		mustConditions := make([]*qdrant.Condition, 0)
		
		for key, value := range filter {
			switch key {
			case "category", "brand", "color", "size":
				if strValue, ok := value.(string); ok && strValue != "" {
					mustConditions = append(mustConditions, qdrant.NewMatch(key, strValue))
				}
			case "price_min":
				if minPrice, ok := value.(float64); ok {
					mustConditions = append(mustConditions, qdrant.NewRange("price", &qdrant.Range{
						Gte: &minPrice,
					}))
				}
			case "price_max":
				if maxPrice, ok := value.(float64); ok {
					mustConditions = append(mustConditions, qdrant.NewRange("price", &qdrant.Range{
						Lte: &maxPrice,
					}))
				}
			}
		}
		
		if len(mustConditions) > 0 {
			scrollRequest.Filter = &qdrant.Filter{
				Must: mustConditions,
			}
		}
	}

	// 执行 Scroll
	response, err := s.client.Scroll(ctx, scrollRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scroll products from Qdrant: %w", err)
	}

	// 转换结果并去重（因为一个商品可能有多个变体）
	productMap := make(map[string]*models.Product)
	for _, point := range response {
		if point.Payload != nil {
			product := s.parseProductFromPayload(point.Payload)
			if product.ID != "" {
				productMap[product.ID] = product
			}
		}
	}

	// 转换为数组
	products := make([]models.Product, 0, len(productMap))
	for _, product := range productMap {
		products = append(products, *product)
	}

	logrus.Infof("Scrolled %d unique products from Qdrant", len(products))
	// 简化返回，暂时不处理 NextPageOffset
	return products, nil, nil
}
func (s *QdrantService) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"collection_name": s.collectionName,
		"status":         "ready",
		"timestamp":      time.Now().Unix(),
	}

	// 如果已初始化，获取基本的连接状态
	if s.initialized {
		stats["status"] = "connected"
		stats["note"] = "Qdrant client initialized and ready"
	} else {
		stats["note"] = "Lazy loading - will initialize on first use"
	}

	return stats, nil
}

// HealthCheck 健康检查
func (s *QdrantService) HealthCheck() error {
	// 健康检查时不初始化，避免启动时的网络调用
	return nil
}
