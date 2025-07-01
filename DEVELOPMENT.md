# 开发指南

## 🚀 快速开始开发模式

### 一键启动开发环境

```bash
# 运行开发脚本（推荐）
./dev.sh
```

这个脚本会自动：
- ✅ 检查 Go 和 Docker 环境
- ✅ 启动 Qdrant 数据库
- ✅ 复制配置文件模板
- ✅ 安装依赖和开发工具
- ✅ 启动热重载开发服务器

### 手动启动步骤

如果你想手动控制每个步骤：

1. **启动 Qdrant 数据库**
```bash
docker run -d --name=qdrant --restart=always \
  -p 6333:6333 -p 6334:6334 \
  -e QDRANT__SERVICE__API_KEY=Axyz.One234 \
  -v qdrant_storage:/qdrant/storage \
  qdrant/qdrant
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，配置你的 API 密钥
```

3. **安装依赖**
```bash
go mod download
```

4. **启动开发服务器**
```bash
# 使用热重载（推荐）
air

# 或者普通模式
export GIN_MODE=debug LOG_LEVEL=debug
go run cmd/server/main.go
```

## 🛠️ 开发工具

### 热重载开发
```bash
# 安装 Air
go install github.com/cosmtrek/air@latest

# 启动热重载
air
```

### 代码质量
```bash
# 代码格式化
go fmt ./...

# 代码检查
golangci-lint run

# 运行测试
go test ./...
```

### API 测试
```bash
# 快速测试所有 API
./quick_test.sh

# 手动测试健康检查
curl http://localhost:8080/api/health
```

## 🔧 开发配置

### 环境变量
开发模式下的重要环境变量：

```bash
# 调试模式
export GIN_MODE=debug
export LOG_LEVEL=debug

# API 配置（必需）
export OPENAI_BASE_URL=https://api.openai.com/v1
export OPENAI_API_KEY=your-api-key-here

# Qdrant 配置
export QDRANT_HOST=localhost
export QDRANT_PORT=6334
export QDRANT_API_KEY=Axyz.One234
```

### 开发端口
- **API 服务**: http://localhost:8080
- **Qdrant UI**: http://localhost:6333/dashboard
- **健康检查**: http://localhost:8080/api/health

## 📝 开发流程

### 1. 功能开发
```bash
# 创建功能分支
git checkout -b feature/new-feature

# 开发过程中使用热重载
air

# 测试功能
./quick_test.sh
```

### 2. 代码提交
```bash
# 格式化代码
go fmt ./...

# 运行测试
go test ./...

# 提交代码
git add .
git commit -m "feat: add new feature"
```

### 3. 调试技巧
```bash
# 查看详细日志
tail -f server.log

# 检查系统状态
curl http://localhost:8080/api/stats

# 测试特定 API
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "测试查询"}'
```

## 🐛 常见问题

### Qdrant 连接问题
```bash
# 检查容器状态
docker ps | grep qdrant

# 重启容器
docker restart qdrant

# 查看容器日志
docker logs qdrant
```

### API 调用失败
```bash
# 检查环境变量
echo $OPENAI_API_KEY
echo $OPENAI_BASE_URL

# 测试 API 连接
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  $OPENAI_BASE_URL/models
```

### 热重载不工作
```bash
# 重新安装 Air
go install github.com/cosmtrek/air@latest

# 检查 .air.toml 配置
cat .air.toml

# 手动启动
go run cmd/server/main.go
```

## 📚 开发资源

### API 文档
- 健康检查: `GET /api/health`
- 系统统计: `GET /api/stats`
- 商品管理: `POST/GET/PUT/DELETE /api/products`
- 智能搜索: `POST /api/search`
- 配置管理: `GET/PUT /api/config/*`

### 项目结构
```
internal/
├── handlers/     # HTTP 处理器
├── services/     # 业务逻辑
├── models/       # 数据模型
└── config/       # 配置管理
```

### 有用的命令
```bash
# 查看所有路由
curl http://localhost:8080/api/stats

# 测试搜索功能
curl -X POST http://localhost:8080/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "蓝色牛仔裤", "limit": 5}'

# 查看配置
curl http://localhost:8080/api/config/function-schema
```

---

Happy Coding! 🎉
