#!/bin/bash

# Search EC2 快速测试脚本
BASE_URL="http://127.0.0.1:8080"

echo "🚀 Search EC2 API 快速测试"
echo "=========================="

# 测试服务器是否运行
echo "1. 检查服务器状态..."
if curl -s --connect-timeout 5 "$BASE_URL/" > /dev/null; then
    echo "✅ 服务器运行正常"
else
    echo "❌ 服务器未运行或无法连接"
    exit 1
fi

# 测试根路径
echo -e "\n2. 测试根路径..."
curl -s -X GET "$BASE_URL/" | head -100

# 测试健康检查
echo -e "\n\n3. 测试健康检查..."
curl -s -X GET "$BASE_URL/api/health" | head -100

# 测试系统统计
echo -e "\n\n4. 测试系统统计..."
curl -s -X GET "$BASE_URL/api/stats" | head -100

# 测试商品创建（应该返回 TODO 消息）
echo -e "\n\n5. 测试商品创建..."
curl -s -X POST "$BASE_URL/api/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商品",
    "category": "测试",
    "price": 100,
    "currency": "CNY"
  }' | head -100

# 测试搜索（应该返回 TODO 消息）
echo -e "\n\n6. 测试商品搜索..."
curl -s -X POST "$BASE_URL/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "测试查询",
    "limit": 5
  }' | head -100

echo -e "\n\n✅ 快速测试完成！"
echo "如需详细测试，请参考 test.md 文件"
