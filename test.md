# Search EC2 API 测试脚本

本文档包含了 Search EC2 自然语言商品搜索系统的完整 API 测试脚本。

## 测试环境
- 服务地址: http://127.0.0.1:8080
- 测试工具: curl

## 1. 基础接口测试

### 1.1 根路径测试
```bash
# 测试根路径
curl -X GET "http://127.0.0.1:8080/" \
  -H "Content-Type: application/json" | jq .
```

### 1.2 健康检查
```bash
# 健康检查
curl -X GET "http://127.0.0.1:8080/api/health" \
  -H "Content-Type: application/json" | jq .
```

### 1.3 系统统计
```bash
# 获取系统统计信息
curl -X GET "http://127.0.0.1:8080/api/stats" \
  -H "Content-Type: application/json" | jq .
```

## 2. 商品管理接口测试

### 2.1 创建单个商品
```bash
# 创建商品
curl -X POST "http://127.0.0.1:8080/api/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "经典蓝色牛仔裤",
    "category": "服装",
    "description": "时尚经典的蓝色牛仔裤，适合日常穿着",
    "price": 299.99,
    "currency": "CNY",
    "brand": "Levi'\''s",
    "color": "蓝色",
    "size": "L",
    "material": "棉质",
    "style": "休闲",
    "gender": "男",
    "occasion": "日常",
    "image_urls": [
      "https://example.com/jeans1.jpg",
      "https://example.com/jeans2.jpg"
    ],
    "tags": ["牛仔裤", "休闲", "经典"],
    "attributes": {
      "wash_type": "水洗",
      "fit": "修身"
    }
  }' | jq .
```

### 2.2 获取商品详情
```bash
# 获取商品详情
curl -X GET "http://127.0.0.1:8080/api/products/test-product-001" \
  -H "Content-Type: application/json" | jq .
```

### 2.3 更新商品信息
```bash
# 更新商品
curl -X PUT "http://127.0.0.1:8080/api/products/test-product-001" \
  -H "Content-Type: application/json" \
  -d '{
    "price": 279.99,
    "description": "时尚经典的蓝色牛仔裤，适合日常穿着，现在特价优惠",
    "tags": ["牛仔裤", "休闲", "经典", "特价"]
  }' | jq .
```

### 2.4 删除商品
```bash
# 删除商品
curl -X DELETE "http://127.0.0.1:8080/api/products/test-product-001" \
  -H "Content-Type: application/json" | jq .
```

### 2.5 批量导入商品
```bash
# 批量导入商品
curl -X POST "http://127.0.0.1:8080/api/products/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "products": [
      {
        "name": "白色T恤",
        "category": "服装",
        "description": "纯棉白色T恤，舒适透气",
        "price": 89.99,
        "currency": "CNY",
        "brand": "Uniqlo",
        "color": "白色",
        "size": "M",
        "material": "纯棉",
        "style": "简约",
        "gender": "中性",
        "occasion": "日常",
        "tags": ["T恤", "纯棉", "基础款"]
      },
      {
        "name": "黑色运动鞋",
        "category": "鞋类",
        "description": "轻便舒适的黑色运动鞋",
        "price": 599.99,
        "currency": "CNY",
        "brand": "Nike",
        "color": "黑色",
        "size": "42",
        "material": "合成材料",
        "style": "运动",
        "gender": "男",
        "occasion": "运动",
        "tags": ["运动鞋", "跑步", "舒适"]
      }
    ]
  }' | jq .
```

### 2.6 重新生成商品变体
```bash
# 重新生成商品变体
curl -X POST "http://127.0.0.1:8080/api/products/test-product-001/variants/regenerate" \
  -H "Content-Type: application/json" \
  -d '{
    "variant_count": 5
  }' | jq .
```

## 3. 搜索接口测试

### 3.1 自然语言搜索
```bash
# 基础搜索
curl -X POST "http://127.0.0.1:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "我想买一条牛仔裤",
    "limit": 10
  }' | jq .
```

```bash
# 带属性的搜索
curl -X POST "http://127.0.0.1:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "我要买蓝色的牛仔裤",
    "limit": 10
  }' | jq .
```

```bash
# 带价格限制的搜索
curl -X POST "http://127.0.0.1:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "买300元以下的牛仔裤",
    "limit": 10
  }' | jq .
```

```bash
# 复合条件搜索
curl -X POST "http://127.0.0.1:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "我要买黑色的Nike运动鞋，500元以下的",
    "limit": 10
  }' | jq .
```

### 3.2 搜索建议
```bash
# 获取搜索建议
curl -X GET "http://127.0.0.1:8080/api/search/suggestions?query=牛仔&limit=5" \
  -H "Content-Type: application/json" | jq .
```

```bash
# 获取更多搜索建议
curl -X GET "http://127.0.0.1:8080/api/search/suggestions?query=运动鞋&limit=8" \
  -H "Content-Type: application/json" | jq .
```

## 4. 配置管理接口测试

### 4.1 Function Calling Schema 管理

#### 获取 Function Schema
```bash
# 获取当前 Function Calling 配置
curl -X GET "http://127.0.0.1:8080/api/config/function-schema" \
  -H "Content-Type: application/json" | jq .
```

#### 更新 Function Schema
```bash
# 更新 Function Calling 配置
curl -X PUT "http://127.0.0.1:8080/api/config/function-schema" \
  -H "Content-Type: application/json" \
  -d '{
    "function_name": "parse_product_query",
    "description": "解析用户商品查询意图，提取商品类型、属性和过滤条件",
    "parameters": {
      "type": "object",
      "properties": {
        "product_type": {
          "type": "string",
          "description": "商品类型或类别，如牛仔裤、T恤、运动鞋等"
        },
        "color": {
          "type": "string",
          "description": "颜色要求，如红色、蓝色、黑色等"
        },
        "price_min": {
          "type": "number",
          "description": "最低价格，单位为元"
        },
        "price_max": {
          "type": "number",
          "description": "最高价格，单位为元"
        },
        "brand": {
          "type": "string",
          "description": "品牌要求，如Nike、Adidas等"
        },
        "size": {
          "type": "string",
          "description": "尺寸要求，如S、M、L、XL等"
        },
        "style": {
          "type": "string",
          "description": "风格要求，如休闲、运动、商务等"
        }
      },
      "required": ["product_type"]
    }
  }' | jq .
```

### 4.2 变体生成提示词管理

#### 获取变体生成提示词
```bash
# 获取当前变体生成提示词
curl -X GET "http://127.0.0.1:8080/api/config/variant-prompt" \
  -H "Content-Type: application/json" | jq .
```

#### 更新变体生成提示词
```bash
# 更新变体生成提示词
curl -X PUT "http://127.0.0.1:8080/api/config/variant-prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "你是一个专业的商品描述生成助手。根据以下商品信息，生成 {variant_count} 个不同的自然语言描述变体。\n\n商品信息：\n- 名称：{product_name}\n- 分类：{category}\n- 颜色：{color}\n- 价格：{price}\n- 品牌：{brand}\n\n生成要求：\n1. 每个变体都要包含核心商品信息\n2. 使用不同的表达方式和同义词\n3. 包含口语化和正式化的表达\n4. 变体之间要有明显差异\n5. 保持信息准确性\n\n请以 JSON 数组格式返回变体列表。"
  }' | jq .
```

## 5. 错误处理测试

### 5.1 无效请求测试
```bash
# 测试无效的商品创建请求（缺少必需字段）
curl -X POST "http://127.0.0.1:8080/api/products" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "缺少名称和分类的商品"
  }' | jq .
```

### 5.2 不存在的资源测试
```bash
# 测试获取不存在的商品
curl -X GET "http://127.0.0.1:8080/api/products/non-existent-product" \
  -H "Content-Type: application/json" | jq .
```

### 5.3 无效的搜索请求
```bash
# 测试空查询搜索
curl -X POST "http://127.0.0.1:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": ""
  }' | jq .
```

## 6. 性能测试

### 6.1 并发搜索测试
```bash
# 并发执行多个搜索请求
for i in {1..5}; do
  curl -X POST "http://127.0.0.1:8080/api/search" \
    -H "Content-Type: application/json" \
    -d "{\"query\": \"测试查询 $i\", \"limit\": 5}" &
done
wait
```

### 6.2 批量操作测试
```bash
# 测试大批量商品导入
curl -X POST "http://127.0.0.1:8080/api/products/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "products": [
      {
        "name": "商品1",
        "category": "测试",
        "price": 100,
        "currency": "CNY"
      },
      {
        "name": "商品2", 
        "category": "测试",
        "price": 200,
        "currency": "CNY"
      },
      {
        "name": "商品3",
        "category": "测试", 
        "price": 300,
        "currency": "CNY"
      }
    ]
  }' | jq .
```

## 7. 完整测试流程脚本

### 7.1 自动化测试脚本
```bash
#!/bin/bash

# Search EC2 API 自动化测试脚本
BASE_URL="http://127.0.0.1:8080"

echo "=== Search EC2 API 测试开始 ==="

# 1. 健康检查
echo "1. 测试健康检查..."
curl -s -X GET "$BASE_URL/api/health" | jq .

# 2. 系统统计
echo "2. 测试系统统计..."
curl -s -X GET "$BASE_URL/api/stats" | jq .

# 3. 创建测试商品
echo "3. 创建测试商品..."
curl -s -X POST "$BASE_URL/api/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试牛仔裤",
    "category": "服装",
    "description": "用于API测试的牛仔裤",
    "price": 199.99,
    "currency": "CNY",
    "brand": "TestBrand",
    "color": "蓝色",
    "size": "M"
  }' | jq .

# 4. 搜索测试
echo "4. 测试商品搜索..."
curl -s -X POST "$BASE_URL/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "蓝色牛仔裤",
    "limit": 5
  }' | jq .

# 5. 配置测试
echo "5. 测试配置获取..."
curl -s -X GET "$BASE_URL/api/config/function-schema" | jq .

echo "=== Search EC2 API 测试完成 ==="
```

### 7.2 保存并运行测试脚本
```bash
# 保存上面的脚本到文件
cat > api_test.sh << 'EOF'
#!/bin/bash
BASE_URL="http://127.0.0.1:8080"
echo "=== Search EC2 API 测试开始 ==="
curl -s -X GET "$BASE_URL/api/health" | jq .
curl -s -X GET "$BASE_URL/api/stats" | jq .
echo "=== Search EC2 API 测试完成 ==="
EOF

# 给脚本执行权限并运行
chmod +x api_test.sh
./api_test.sh
```

## 8. 测试数据清理

### 8.1 清理测试商品
```bash
# 删除测试过程中创建的商品
curl -X DELETE "http://127.0.0.1:8080/api/products/test-product-001" \
  -H "Content-Type: application/json" | jq .
```

## 注意事项

1. **jq 工具**: 测试脚本使用 `jq` 来格式化 JSON 输出，如果没有安装，可以去掉 `| jq .` 部分
2. **并发测试**: 并发测试可能会对服务器造成压力，请根据实际情况调整
3. **数据持久化**: 测试数据会保存在系统中，记得及时清理
4. **错误处理**: 某些 API 可能返回错误，这是正常的测试行为
5. **响应时间**: 首次调用某些 API 可能较慢，因为需要初始化服务

## 预期响应格式

所有 API 响应都遵循统一格式：
```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "timestamp": 1672531200
}
```

错误响应格式：
```json
{
  "code": 400,
  "message": "error message",
  "timestamp": 1672531200
}
```
