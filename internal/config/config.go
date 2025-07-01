package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Qdrant   QdrantConfig   `mapstructure:"qdrant"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Search   SearchConfig   `mapstructure:"search"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Features FeaturesConfig `mapstructure:"features"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// QdrantConfig Qdrant 配置
type QdrantConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	APIKey         string `mapstructure:"api_key"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     int    `mapstructure:"vector_size"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	ChatModel      string `mapstructure:"chat_model"`
	MaxTokens      int    `mapstructure:"max_tokens"`
	Timeout        int    `mapstructure:"timeout"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	MaxResults          int     `mapstructure:"max_results"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold"`
	EnableCache         bool    `mapstructure:"enable_cache"`
	CacheTTL            int     `mapstructure:"cache_ttl"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// FeaturesConfig 功能开关配置
type FeaturesConfig struct {
	EnableBatchImport       bool `mapstructure:"enable_batch_import"`
	EnableVariantGeneration bool `mapstructure:"enable_variant_generation"`
	EnableFunctionCalling   bool `mapstructure:"enable_function_calling"`
	EnableSearchSuggestions bool `mapstructure:"enable_search_suggestions"`
}

// FunctionCallingSchema Function Calling 配置结构
type FunctionCallingSchema struct {
	FunctionName string                 `json:"function_name"`
	Description  string                 `json:"description"`
	Parameters   map[string]interface{} `json:"parameters"`
}

var (
	AppConfig             *Config
	FunctionSchema        *FunctionCallingSchema
	VariantPromptTemplate string
)

// Load 加载配置
func Load() error {
	// 设置配置文件路径
	viper.SetConfigName("app_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// 设置环境变量前缀
	viper.SetEnvPrefix("SEARCH_EC2")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 从环境变量覆盖敏感配置
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		AppConfig.OpenAI.APIKey = apiKey
	}

	// 加载 Function Calling Schema
	if err := loadFunctionCallingSchema(); err != nil {
		return fmt.Errorf("failed to load function calling schema: %w", err)
	}

	// 加载变体生成提示词模板
	if err := loadVariantPromptTemplate(); err != nil {
		return fmt.Errorf("failed to load variant prompt template: %w", err)
	}

	return nil
}

// loadFunctionCallingSchema 加载 Function Calling 配置
func loadFunctionCallingSchema() error {
	schemaPath := findConfigFile("function_calling_schema.json")
	if schemaPath == "" {
		return fmt.Errorf("function_calling_schema.json not found")
	}

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read function calling schema: %w", err)
	}

	FunctionSchema = &FunctionCallingSchema{}
	if err := json.Unmarshal(data, FunctionSchema); err != nil {
		return fmt.Errorf("failed to parse function calling schema: %w", err)
	}

	return nil
}

// loadVariantPromptTemplate 加载变体生成提示词模板
func loadVariantPromptTemplate() error {
	promptPath := findConfigFile("variant_prompt.txt")
	if promptPath == "" {
		return fmt.Errorf("variant_prompt.txt not found")
	}

	data, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("failed to read variant prompt template: %w", err)
	}

	VariantPromptTemplate = string(data)
	return nil
}

// findConfigFile 查找配置文件
func findConfigFile(filename string) string {
	paths := []string{
		filepath.Join("config", filename),
		filepath.Join("../config", filename),
		filepath.Join("../../config", filename),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetAddress 获取服务器地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetQdrantAddress 获取 Qdrant 地址
func (c *Config) GetQdrantAddress() string {
	return fmt.Sprintf("http://%s:%d", c.Qdrant.Host, c.Qdrant.Port)
}

// Reload 重新加载配置（用于热加载）
func Reload() error {
	return Load()
}
