package models

import (
	"time"
)

// Product 商品数据结构
type Product struct {
	ID          string                 `json:"id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Category    string                 `json:"category" binding:"required"`
	Description string                 `json:"description"`
	Price       float64                `json:"price" binding:"min=0"`
	Currency    string                 `json:"currency" binding:"required"`
	Brand       string                 `json:"brand"`
	Color       string                 `json:"color"`
	Size        string                 `json:"size"`
	Material    string                 `json:"material"`
	Style       string                 `json:"style"`
	Gender      string                 `json:"gender"`
	Occasion    string                 `json:"occasion"`
	ImageURLs   []string               `json:"image_urls"`
	Tags        []string               `json:"tags"`
	Attributes  map[string]interface{} `json:"attributes"` // 动态字段
	Status      string                 `json:"status"`     // active, inactive, deleted
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ProductVariant 商品变体结构
type ProductVariant struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Text        string    `json:"text"`        // 变体描述文本
	Vector      []float32 `json:"vector"`      // 向量表示
	GeneratedAt time.Time `json:"generated_at"`
}

// ProductCreateRequest 创建商品请求
type ProductCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Category    string                 `json:"category" binding:"required"`
	Description string                 `json:"description"`
	Price       float64                `json:"price" binding:"min=0"`
	Currency    string                 `json:"currency" binding:"required"`
	Brand       string                 `json:"brand"`
	Color       string                 `json:"color"`
	Size        string                 `json:"size"`
	Material    string                 `json:"material"`
	Style       string                 `json:"style"`
	Gender      string                 `json:"gender"`
	Occasion    string                 `json:"occasion"`
	ImageURLs   []string               `json:"image_urls"`
	Tags        []string               `json:"tags"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// ProductUpdateRequest 更新商品请求
type ProductUpdateRequest struct {
	Name        *string                `json:"name"`
	Category    *string                `json:"category"`
	Description *string                `json:"description"`
	Price       *float64               `json:"price"`
	Currency    *string                `json:"currency"`
	Brand       *string                `json:"brand"`
	Color       *string                `json:"color"`
	Size        *string                `json:"size"`
	Material    *string                `json:"material"`
	Style       *string                `json:"style"`
	Gender      *string                `json:"gender"`
	Occasion    *string                `json:"occasion"`
	ImageURLs   []string               `json:"image_urls"`
	Tags        []string               `json:"tags"`
	Attributes  map[string]interface{} `json:"attributes"`
	Status      *string                `json:"status"`
}

// BatchImportRequest 批量导入请求
type BatchImportRequest struct {
	Products []ProductCreateRequest `json:"products" binding:"required,min=1"`
}

// BatchImportResponse 批量导入响应
type BatchImportResponse struct {
	Total     int                    `json:"total"`
	Success   int                    `json:"success"`
	Failed    int                    `json:"failed"`
	Errors    []BatchImportError     `json:"errors,omitempty"`
	ProcessID string                 `json:"process_id"`
}

// BatchImportError 批量导入错误
type BatchImportError struct {
	Index   int    `json:"index"`
	Product string `json:"product"`
	Error   string `json:"error"`
}

// ToProduct 将创建请求转换为商品对象
func (req *ProductCreateRequest) ToProduct() *Product {
	now := time.Now()
	return &Product{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Price:       req.Price,
		Currency:    req.Currency,
		Brand:       req.Brand,
		Color:       req.Color,
		Size:        req.Size,
		Material:    req.Material,
		Style:       req.Style,
		Gender:      req.Gender,
		Occasion:    req.Occasion,
		ImageURLs:   req.ImageURLs,
		Tags:        req.Tags,
		Attributes:  req.Attributes,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ApplyUpdate 应用更新请求
func (p *Product) ApplyUpdate(req *ProductUpdateRequest) {
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Currency != nil {
		p.Currency = *req.Currency
	}
	if req.Brand != nil {
		p.Brand = *req.Brand
	}
	if req.Color != nil {
		p.Color = *req.Color
	}
	if req.Size != nil {
		p.Size = *req.Size
	}
	if req.Material != nil {
		p.Material = *req.Material
	}
	if req.Style != nil {
		p.Style = *req.Style
	}
	if req.Gender != nil {
		p.Gender = *req.Gender
	}
	if req.Occasion != nil {
		p.Occasion = *req.Occasion
	}
	if req.ImageURLs != nil {
		p.ImageURLs = req.ImageURLs
	}
	if req.Tags != nil {
		p.Tags = req.Tags
	}
	if req.Attributes != nil {
		p.Attributes = req.Attributes
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	p.UpdatedAt = time.Now()
}
