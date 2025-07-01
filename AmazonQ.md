# 自然语言商品搜索系统规划讨论

## 系统概述
- **目标**: 构建一个基于向量的通用商品检索系统
- **特点**: 完全解耦，独立于电商系统
- **输入方式**: 支持用户录入和API批量插入

## 需求分析

### 用户查询类型
- 基础商品查询: "我想买一个牛仔裤"
- 属性限定查询: "我要买黄色的牛仔裤"
- 价格限定查询: "买100元以下的牛仔裤"
- 多维度组合查询: 颜色+价格+商品类型

### 技术方案选择

#### 向量化策略对比
1. **单一混合向量**: 所有信息编码成一个综合向量
   - 优点: 实现简单
   - 缺点: 召回率相对较低，精确属性匹配效果差

2. **多向量融合**: 分别生成不同维度向量后融合
   - 优点: 能处理多维度信息
   - 缺点: 向量融合可能丢失精确性

3. **混合检索**: 向量检索 + 传统过滤器结合 ✅
   - 优点: 召回率最高，语义匹配+精确过滤
   - 缺点: 计算复杂度稍高
   - **最终选择**: 适合问答式检索场景，允许后台思考时间

## 数据结构设计

### 商品变体向量存储方案
- **1对多向量存储**: 一个商品对应多个变体表述向量
- **字段设计**: 固定字段 + 动态字段结合
- **变体示例**: 
  - 商品ID: 001
  - 变体1: "蓝色牛仔裤 32码" → 向量A
  - 变体2: "深蓝色丹宁裤 L码" → 向量B
  - 变体3: "休闲牛仔长裤 大码" → 向量C

### 商品变体生成策略 ✅
- **AI辅助生成**: 通过提示词让AI根据商品信息生成多种自然语言变体
- **优点**: 自动化程度高，能生成丰富的语言表达变体
- **实现方式**: 商品信息 → 提示词模板 → AI生成变体 → 向量化存储
- **变体数量**: 通过提示词灵活控制，适应不同业务需求

## 技术栈选择

### 确定的技术栈 ✅
- **后端框架**: Golang + Gin
- **向量数据库**: Qdrant
- **向量化模型**: OpenAI格式（支持灵活更换）
- **部署方式**: Docker容器化

### 数据存储决策 ✅
- **只使用Qdrant**: 更新不频繁，专注查询效率
- **数据存储**: 商品信息存储在Qdrant的payload中
- **架构优势**: 简化系统复杂度，减少数据同步开销

### Qdrant集合设计 ✅
- **单Collection方案**: 所有商品变体存储在一个Collection中
- **数据结构**: 
  - 向量: 商品变体的embedding
  - Payload: 商品ID + 变体文本 + 商品完整信息
- **优势**: 一次查询获得所有信息，查询效率最高

### 混合检索实现方案 ✅
- **查询解析**: 使用Function Calling解析用户意图
  - 输入: "我要买100元以下的黄色牛仔裤"
  - 解析结果: {product_type: "牛仔裤", color: "黄色", max_price: 100}
- **检索流程**: 
  1. 向量检索: 使用product_type进行语义搜索
  2. 过滤器: 使用解析出的结构化条件过滤
- **优势**: 精确解析 + 高召回率的完美结合
- **动态配置**: Function Calling schema可配置化，Qdrant支持灵活payload

## API接口设计

### 商品管理接口
- **POST /api/products** - 单个商品录入
- **POST /api/products/batch** - 批量商品导入
- **PUT /api/products/{id}** - 商品更新
- **DELETE /api/products/{id}** - 商品删除
- **GET /api/products/{id}** - 获取商品详情

### 搜索接口
- **POST /api/search** - 自然语言搜索
  - 输入: 用户查询文本
  - 输出: 商品列表 + 相关度评分 + 匹配原因
- **GET /api/search/suggestions** - 搜索建议/自动补全

### 配置管理接口
- **GET /api/config/function-schema** - 获取Function Calling配置
- **PUT /api/config/function-schema** - 更新Function Calling配置
- **GET /api/config/variant-prompt** - 获取变体生成提示词
- **PUT /api/config/variant-prompt** - 更新变体生成提示词

### 系统管理接口
- **POST /api/products/{id}/variants/regenerate** - 重新生成商品变体
- **GET /api/health** - 健康检查
- **GET /api/stats** - 系统统计信息

## 开发任务拆分

### 阶段1: 基础架构
1. 项目初始化和依赖管理
2. Gin框架搭建和路由设计
3. Qdrant连接和基础操作封装
4. 配置文件管理系统

### 阶段2: 核心功能
1. 商品数据模型设计
2. AI变体生成服务
3. 向量化服务集成
4. Function Calling解析服务

### 阶段3: 检索系统
1. 混合检索算法实现
2. 搜索接口开发
3. 结果排序和评分机制
4. 搜索性能优化

### 阶段4: 管理功能
1. 商品CRUD接口实现
2. 批量导入功能
3. 配置管理接口
4. 系统监控和统计

### 阶段5: 部署和优化
1. Docker容器化
2. 性能测试和优化
3. 错误处理和日志系统
4. API文档生成

### 待确定问题
- Function Calling的具体schema设计
- AI变体生成的提示词设计

## 配置管理方案

### 动态配置存储策略 ✅
**选择方案：配置文件存储**
- JSON/YAML文件存储schema配置
- 支持版本控制和备份
- 重启服务生效或支持热加载
- 配置和代码分离，便于管理

### 配置文件结构设计 ✅
```
config/
├── function_calling_schema.json  # Function Calling配置
├── variant_prompt.txt           # 变体生成提示词
└── app_config.yaml             # 应用基础配置
```

### Function Calling Schema示例
```json
{
  "function_name": "parse_product_query",
  "description": "解析用户商品查询意图",
  "parameters": {
    "type": "object",
    "properties": {
      "product_type": {
        "type": "string",
        "description": "商品类型，如牛仔裤、T恤等"
      },
      "color": {
        "type": "string", 
        "description": "颜色要求"
      },
      "price_min": {
        "type": "number",
        "description": "最低价格"
      },
      "price_max": {
        "type": "number", 
        "description": "最高价格"
      },
      "brand": {
        "type": "string",
        "description": "品牌要求"
      },
      "size": {
        "type": "string",
        "description": "尺寸要求"
      }
    }
  }
}
```
