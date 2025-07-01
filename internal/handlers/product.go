package handlers

import (
	"fmt"
	"search-ec2/internal/models"
	"search-ec2/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProductHandler 商品处理器
type ProductHandler struct {
	serviceManager *services.ServiceManager
}

// NewProductHandler 创建商品处理器
func NewProductHandler(serviceManager *services.ServiceManager) *ProductHandler {
	return &ProductHandler{
		serviceManager: serviceManager,
	}
}

// CreateProduct 创建商品
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// 转换为商品对象
	product := req.ToProduct()
	product.ID = uuid.New().String()

	logrus.Infof("Creating product: %s", product.Name)

	// 生成商品变体
	variants, err := h.serviceManager.VariantGeneration.GenerateVariantsWithEmbeddings(
		product, 5, h.serviceManager.Embedding,
	)
	if err != nil {
		logrus.Errorf("Failed to generate variants: %v", err)
		InternalErrorResponse(c, "Failed to generate product variants")
		return
	}

	// 插入到 Qdrant
	if err := h.serviceManager.Qdrant.InsertProduct(product, variants); err != nil {
		logrus.Errorf("Failed to insert product: %v", err)
		InternalErrorResponse(c, "Failed to save product")
		return
	}

	logrus.Infof("Product created successfully: %s with %d variants", product.ID, len(variants))

	response := map[string]interface{}{
		"product_id":     product.ID,
		"name":          product.Name,
		"variants_count": len(variants),
		"created_at":    product.CreatedAt,
	}

	SuccessResponse(c, response)
}

// GetProduct 获取商品详情
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		BadRequestResponse(c, "Product ID is required")
		return
	}

	product, err := h.serviceManager.Qdrant.GetProduct(productID)
	if err != nil {
		logrus.Errorf("Failed to get product %s: %v", productID, err)
		NotFoundResponse(c, "Product not found")
		return
	}

	SuccessResponse(c, product)
}

// UpdateProduct 更新商品
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		BadRequestResponse(c, "Product ID is required")
		return
	}

	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// 获取现有商品
	product, err := h.serviceManager.Qdrant.GetProduct(productID)
	if err != nil {
		logrus.Errorf("Failed to get product %s: %v", productID, err)
		NotFoundResponse(c, "Product not found")
		return
	}

	// 应用更新
	product.ApplyUpdate(&req)

	// 重新生成变体
	variants, err := h.serviceManager.VariantGeneration.GenerateVariantsWithEmbeddings(
		product, 5, h.serviceManager.Embedding,
	)
	if err != nil {
		logrus.Errorf("Failed to regenerate variants: %v", err)
		InternalErrorResponse(c, "Failed to update product variants")
		return
	}

	// 删除旧数据并插入新数据
	if err := h.serviceManager.Qdrant.DeleteProduct(productID); err != nil {
		logrus.Errorf("Failed to delete old product data: %v", err)
	}

	if err := h.serviceManager.Qdrant.InsertProduct(product, variants); err != nil {
		logrus.Errorf("Failed to insert updated product: %v", err)
		InternalErrorResponse(c, "Failed to save updated product")
		return
	}

	logrus.Infof("Product updated successfully: %s", productID)

	response := map[string]interface{}{
		"product_id":     product.ID,
		"name":          product.Name,
		"variants_count": len(variants),
		"updated_at":    product.UpdatedAt,
	}

	SuccessResponse(c, response)
}

// DeleteProduct 删除商品
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		BadRequestResponse(c, "Product ID is required")
		return
	}

	if err := h.serviceManager.Qdrant.DeleteProduct(productID); err != nil {
		logrus.Errorf("Failed to delete product %s: %v", productID, err)
		InternalErrorResponse(c, "Failed to delete product")
		return
	}

	logrus.Infof("Product deleted successfully: %s", productID)

	response := map[string]interface{}{
		"product_id": productID,
		"deleted_at": time.Now(),
	}

	SuccessResponse(c, response)
}

// BatchImport 批量导入商品
func (h *ProductHandler) BatchImport(c *gin.Context) {
	var req models.BatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	processID := uuid.New().String()
	logrus.Infof("Starting batch import with process ID: %s, total products: %d", processID, len(req.Products))

	response := models.BatchImportResponse{
		Total:     len(req.Products),
		Success:   0,
		Failed:    0,
		Errors:    []models.BatchImportError{},
		ProcessID: processID,
	}

	// 逐个处理商品
	for i, productReq := range req.Products {
		product := productReq.ToProduct()
		product.ID = uuid.New().String()

		// 生成变体
		variants, err := h.serviceManager.VariantGeneration.GenerateVariantsWithEmbeddings(
			product, 3, h.serviceManager.Embedding, // 批量导入时减少变体数量
		)
		if err != nil {
			logrus.Errorf("Failed to generate variants for product %d: %v", i, err)
			response.Failed++
			response.Errors = append(response.Errors, models.BatchImportError{
				Index:   i,
				Product: product.Name,
				Error:   fmt.Sprintf("Failed to generate variants: %v", err),
			})
			continue
		}

		// 插入商品
		if err := h.serviceManager.Qdrant.InsertProduct(product, variants); err != nil {
			logrus.Errorf("Failed to insert product %d: %v", i, err)
			response.Failed++
			response.Errors = append(response.Errors, models.BatchImportError{
				Index:   i,
				Product: product.Name,
				Error:   fmt.Sprintf("Failed to save product: %v", err),
			})
			continue
		}

		response.Success++
		logrus.Debugf("Successfully imported product %d: %s", i, product.Name)
	}

	logrus.Infof("Batch import completed: %d success, %d failed", response.Success, response.Failed)
	SuccessResponse(c, response)
}

// RegenerateVariants 重新生成商品变体
func (h *ProductHandler) RegenerateVariants(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		BadRequestResponse(c, "Product ID is required")
		return
	}

	var req models.VariantGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用默认值
		req.VariantCount = 5
	}

	if req.VariantCount <= 0 || req.VariantCount > 20 {
		req.VariantCount = 5
	}

	err := h.serviceManager.VariantGeneration.RegenerateVariants(
		productID, req.VariantCount, h.serviceManager.Qdrant, h.serviceManager.Embedding,
	)
	if err != nil {
		logrus.Errorf("Failed to regenerate variants for product %s: %v", productID, err)
		InternalErrorResponse(c, "Failed to regenerate variants")
		return
	}

	logrus.Infof("Variants regenerated successfully for product: %s", productID)

	response := map[string]interface{}{
		"product_id":     productID,
		"variant_count":  req.VariantCount,
		"regenerated_at": time.Now(),
	}

	SuccessResponse(c, response)
}
