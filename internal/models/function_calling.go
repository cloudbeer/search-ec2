package models

// OpenAIRequest OpenAI API 请求结构
type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIMessage        `json:"messages"`
	Functions   []OpenAIFunction       `json:"functions,omitempty"`
	FunctionCall interface{}           `json:"function_call,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
}

// OpenAIMessage OpenAI 消息结构
type OpenAIMessage struct {
	Role         string                 `json:"role"`
	Content      string                 `json:"content,omitempty"`
	FunctionCall *OpenAIFunctionCall    `json:"function_call,omitempty"`
}

// OpenAIFunction OpenAI Function 定义
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAIFunctionCall OpenAI Function Call
type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAIResponse OpenAI API 响应结构
type OpenAIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []OpenAIChoice   `json:"choices"`
	Usage   OpenAIUsage      `json:"usage"`
	Error   *OpenAIError     `json:"error,omitempty"`
}

// OpenAIChoice OpenAI 选择结构
type OpenAIChoice struct {
	Index        int                `json:"index"`
	Message      OpenAIMessage      `json:"message"`
	FinishReason string             `json:"finish_reason"`
}

// OpenAIUsage OpenAI 使用统计
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError OpenAI 错误结构
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// EmbeddingRequest 向量化请求
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse 向量化响应
type EmbeddingResponse struct {
	Object string           `json:"object"`
	Data   []EmbeddingData  `json:"data"`
	Model  string           `json:"model"`
	Usage  EmbeddingUsage   `json:"usage"`
	Error  *OpenAIError     `json:"error,omitempty"`
}

// EmbeddingData 向量数据
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingUsage 向量化使用统计
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// VariantGenerationRequest 变体生成请求
type VariantGenerationRequest struct {
	Product      *Product `json:"product" binding:"required"`
	VariantCount int      `json:"variant_count" binding:"min=1,max=20"`
}

// VariantGenerationResponse 变体生成响应
type VariantGenerationResponse struct {
	ProductID string   `json:"product_id"`
	Variants  []string `json:"variants"`
	Count     int      `json:"count"`
}
