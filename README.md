# Search EC2 - 自然语言商品搜索系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![API Status](https://img.shields.io/badge/API-RESTful-orange.svg)](docs/api.md)

一个基于向量数据库的智能商品搜索系统，支持自然语言查询、语义理解和多维度过滤。

## 🌟 核心特性

### 🧠 智能搜索
- **自然语言理解**: 支持"我想买一条蓝色的牛仔裤"等自然语言查询
- **语义相似度匹配**: 基于向量相似度的智能搜索
- **意图解析**: 使用 Function Calling 自动解析查询意图和过滤条件

### 🎯 精确过滤
- **多维度过滤**: 支持价格、品牌、颜色、尺寸等多种过滤条件
- **复杂逻辑**: 支持 Must/Should/MustNot 组合过滤
- **范围查询**: 价格区间、时间范围等灵活查询

### 🤖 AI 驱动
- **自动变体生成**: LLM 自动为商品生成多种自然语言描述变体
- **向量化存储**: 自动将商品信息转换为高维向量
- **智能匹配**: 基于语义相似度的智能商品推荐

## 🚀 快速开始

### 1. 环境要求
- Go 1.21+
- Docker (用于 Qdrant)
- OpenAI 兼容 API

### 2. 启动开发环境
```bash
# 一键启动开发环境
./dev.sh
```

### 3. 手动启动
```bash
# 启动 Qdrant
docker run -d --name=qdrant -p 6333:6333 -p 6334:6334 \
  -e QDRANT__SERVICE__API_KEY=Axyz.One234 \
  qdrant/qdrant

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置 API 密钥

# 启动服务
go run cmd/server/main.go
```

## 📖 API 使用

### 创建商品
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "经典牛仔裤",
    "category": "服装",
    "price": 299.99,
    "brand": "Levi'\''s",
    "color": "蓝色"
  }'
```

### 智能搜索
```bash
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "我想买一条蓝色的牛仔裤",
    "limit": 10
  }'
```

## 🏗️ 系统架构

- **RESTful API**: 基于 Gin 框架
- **向量数据库**: Qdrant 高性能向量存储
- **AI 服务**: OpenAI 兼容的 LLM 和 Embedding
- **智能解析**: Function Calling 意图理解
- **自动变体**: LLM 生成商品描述变体

## 📁 项目结构

```
search-ec2/
├── cmd/server/          # 应用入口
├── internal/
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑
│   ├── models/          # 数据模型
│   └── config/          # 配置管理
├── config/              # 配置文件
├── .env.example         # 环境变量模板
└── dev.sh              # 开发脚本
```

## 🛠️ 开发

```bash
# 热重载开发
air

# 运行测试
./quick_test.sh

# 代码格式化
go fmt ./...
```

## 📊 功能特性

- ✅ 自然语言商品搜索
- ✅ 多维度过滤和排序
- ✅ AI 自动生成商品变体
- ✅ 向量化语义匹配
- ✅ RESTful API 接口
- ✅ 实时热重载开发
- ✅ Docker 容器化部署

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📝 许可证

MIT License
