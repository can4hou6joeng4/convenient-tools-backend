# 便捷工具集 - 后端服务

[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/can4hou6joeng4/convenient-tools-project-v1-backend/blob/main/LICENSE)
[![Platform](https://img.shields.io/badge/platform-Go-00ADD8.svg)](https://golang.org)
[![API-Powered](https://img.shields.io/badge/API--Powered-Yes-orange.svg)](https://your-online-address)

基于Go语言和Fiber框架构建的高性能后端服务，为便捷工具集应用提供强大的API支持。集成视频解析、媒体代理、文件处理等多种实用功能，专为现代Web应用和微信小程序设计。

> 🔥 **特别说明**：本项目是一个面向现代Web开发的开源后端服务，展示了如何使用Go语言构建高性能、可扩展的API服务。从系统架构设计、RESTful API开发到容器化部署，提供完整的后端开发最佳实践。无论你是Go语言新手还是有经验的后端开发者，都能从中学习现代后端服务的设计理念和实现方法。

## 📋 功能特点

- 🎬 **视频解析服务** - 支持抖音、小红书、西瓜视频等主流平台链接解析
- 🌐 **媒体代理服务** - 解决微信小程序跨域访问限制，支持视频格式转换
- 📄 **文件处理服务** - PDF转换、文件上传等文档处理功能
- 🔧 **工具管理系统** - 动态工具注册与管理
- 🚀 **高性能架构** - 基于Fiber框架，支持高并发请求处理
- 📝 **完整API文档** - 集成Swagger自动生成API文档

## 🏗️ 系统架构

- **Web框架**：Go Fiber (高性能HTTP框架)
- **数据库**：PostgreSQL (关系型数据库)
- **缓存**：Redis (内存数据库)
- **对象存储**：腾讯云COS (文件存储服务)
- **容器化**：Docker & Docker Compose
- **API文档**：Swagger/OpenAPI 3.0

## 🛠️ 技术栈

- **编程语言**：🐹 Go 1.24+
- **Web框架**：⚡ [Fiber v2.52.6](https://github.com/gofiber/fiber) - 高性能Web框架
- **数据库ORM**：🔗 GORM - Go语言ORM库
- **数据库**：🐘 PostgreSQL 15+ - 企业级关系型数据库
- **缓存系统**：🔴 Redis 7+ - 高性能内存数据库
- **文档生成**：📋 Swagger - 自动API文档生成
- **容器化**：🐳 Docker & Docker Compose - 容器化部署
- **热重载**：🔥 Air - 开发环境热重载工具
- **HTTP客户端**：🌐 Resty - Go语言HTTP客户端
- **JSON处理**：⚡ gjson - 高性能JSON解析库

## 🚀 快速开始

### 环境要求

- Go 1.24 或更高版本
- Docker 和 Docker Compose
- PostgreSQL 15+ (可选，可使用Docker)
- Redis 7+ (可选，可使用Docker)

### 本地开发环境配置

1. **克隆项目**

```bash
git clone https://github.com/can4hou6joeng4/convenient-tools-project-v1-backend.git
cd convenient-tools-project-v1-backend
```

2. **环境变量配置**

复制环境变量模板并配置：

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量文件
vim .env
```

环境变量配置示例：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=convenient_tools

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 腾讯云COS配置
COS_SECRET_ID=your_secret_id
COS_SECRET_KEY=your_secret_key
COS_BUCKET_URL=your_bucket_url

# 服务配置
SERVER_PORT=8082
```

3. **安装依赖**

```bash
go mod tidy
```

4. **启动开发环境**

使用Docker Compose启动所有依赖服务：

```bash
# 启动数据库和Redis
make start

# 或手动启动
docker-compose up -d
```

使用Air进行热重载开发：

```bash
# 安装Air (如果未安装)
go install github.com/cosmtrek/air@latest

# 启动热重载开发
air
```

### 生产环境部署

#### Docker部署方式

1. **构建Docker镜像**

```bash
# 构建生产镜像
docker build -t convenient-tools-backend .
```

2. **使用Docker Compose部署**

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d
```

#### 传统部署方式

1. **编译应用**

```bash
# 编译生产版本
go build -o bin/server cmd/api/main.go
```

2. **启动服务**

```bash
# 设置环境变量
export ENV=production

# 启动服务
./bin/server
```

## 📚 API文档

### 在线文档

启动服务后，可通过以下地址访问API文档：

- **本地环境**: http://localhost:8082/swagger/
- **生产环境**: https://your-online-address/swagger/

### 主要API接口

| 接口路径 | 方法 | 功能描述 |
|---------|------|----------|
| `/api/tools/parse` | POST | 视频分享链接解析 |
| `/api/tools/media-proxy` | GET | 媒体资源代理服务 |
| `/api/tools/list` | GET | 获取工具列表 |
| `/api/tools` | POST | 创建新工具 |
| `/api/file/upload` | POST | 文件上传服务 |

### API使用示例

#### 视频解析接口

**请求示例**：

```bash
curl -X POST "https://your-online-address/api/tools/parse" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://v.douyin.com/iFRvNakm/"
  }'
```

**响应示例**：

```json
{
  "status": "success",
  "message": "Parse URL success",
  "data": {
    "title": "视频标题",
    "video_url": "https://example.com/video.mp4",
    "cover_url": "https://example.com/cover.jpg",
    "author": {
      "name": "作者名称",
      "avatar": "https://example.com/avatar.jpg",
      "uid": "123456"
    },
    "images": []
  }
}
```

#### 媒体代理接口

**请求示例**：

```bash
curl -X GET "https://your-online-address/api/tools/media-proxy?url=https://example.com/video.mp4&type=video"
```

## 🚀 核心功能详解

### 🎬 视频解析服务

支持主流视频平台的分享链接解析：

- **抖音** (v.douyin.com, www.douyin.com)
- **小红书** (xiaohongshu.com, xhslink.com) 
- **西瓜视频** (v.ixigua.com)

**功能特性**：
- 自动识别平台类型
- 提取视频直链地址
- 获取视频封面和作者信息
- 支持图集内容解析
- 智能URL正则匹配

### 🌐 媒体代理服务

专为微信小程序设计的媒体资源代理服务：

**核心功能**：
- 跨域资源访问代理
- 视频格式自动转换 (支持MP4转换)
- 图片资源代理优化
- 流式数据传输
- 智能重试机制

**技术特点**：
- 支持FFmpeg视频转换
- 小红书图片特殊处理
- 自动Content-Type检测
- 缓存控制优化

### 📄 文件处理服务

提供完整的文件上传和处理功能：

- PDF文件上传和存储
- 腾讯云COS集成
- 文件类型验证
- 安全文件处理

## 🔮 开发路线图

### 即将推出的功能

- 📝 **文本处理工具** - 编码转换、格式化、内容分析
- 🌐 **网页截图服务** - URL转图片功能
- 🔍 **OCR文字识别** - 图片文字提取
- 🔑 **JWT认证系统** - API访问控制

### 性能优化计划

- 🚀 **缓存策略优化** - Redis缓存层增强
- 📈 **并发处理提升** - Go协程池优化
- 🔧 **API限流机制** - 防止接口滥用
- 📊 **监控告警系统** - 服务状态实时监控

## 💻 开发指南

### 项目结构

```
convenient-tools-project-v1-backend/
├── cmd/                    # 应用程序入口
│   └── api/               # API服务入口
├── config/                # 配置管理
├── db/                    # 数据库连接和迁移
├── docs/                  # Swagger文档
├── handlers/              # HTTP请求处理器
├── models/                # 数据模型定义
├── repositories/          # 数据访问层
├── utils/                 # 工具函数库
├── .air.toml             # Air热重载配置
├── .gitignore            # Git忽略文件
├── docker-compose.yaml   # Docker编排配置
├── Dockerfile            # Docker构建文件
├── go.mod                # Go模块依赖
├── go.sum                # 依赖版本锁定
└── Makefile              # 构建脚本
```

### 开发规范

1. **代码风格**
   - 遵循Go官方代码规范
   - 使用gofmt格式化代码
   - 添加必要的注释和文档

2. **API设计**
   - 遵循RESTful设计原则
   - 统一的响应格式
   - 完整的错误处理

3. **数据库设计**
   - 使用GORM进行数据库操作
   - 合理的表结构设计
   - 数据库迁移管理

### 本地开发流程

1. **环境准备**

```bash
# 安装Go依赖
go mod download

# 启动依赖服务
docker-compose up -d postgres redis
```

2. **数据库初始化**

```bash
# 运行数据库迁移
go run cmd/api/main.go migrate
```

3. **启动开发服务**

```bash
# 热重载开发
air

# 或直接运行
go run cmd/api/main.go
```

4. **API测试**

```bash
# 健康检查
curl http://localhost:8082/health

# 查看API文档
open http://localhost:8082/swagger/
```

## 🚢 部署指南

### Docker部署

1. **单容器部署**

```bash
# 构建镜像
docker build -t convenient-tools-backend .

# 运行容器
docker run -d \
  --name convenient-tools-backend \
  -p 8082:8082 \
  -e DB_HOST=your_db_host \
  -e REDIS_HOST=your_redis_host \
  convenient-tools-backend
```

2. **Docker Compose部署**

```bash
# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 云服务器部署

1. **系统要求**
   - Ubuntu 20.04+ / CentOS 8+
   - 2GB+ RAM
   - 20GB+ 存储空间

2. **安装依赖**

```bash
# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

3. **部署应用**

```bash
# 克隆代码
git clone https://github.com/can4hou6joeng4/convenient-tools-project-v1-backend.git
cd convenient-tools-project-v1-backend

# 配置环境变量
cp .env.example .env
vim .env

# 启动服务
docker-compose up -d
```

### Nginx反向代理配置

```nginx
server {
    listen 80;
    server_name your-online-address;

    location /api/ {
        proxy_pass http://localhost:8082/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /swagger/ {
        proxy_pass http://localhost:8082/swagger/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 📊 性能监控

### 系统监控指标

- **响应时间**: API接口平均响应时间
- **并发处理**: 同时处理的请求数量
- **错误率**: API调用失败率统计
- **资源使用**: CPU、内存、磁盘使用情况

### 日志管理

```bash
# 查看应用日志
docker-compose logs -f backend

# 查看错误日志
docker-compose logs -f backend | grep ERROR

# 日志轮转配置
# 在docker-compose.yml中配置logging选项
```

## 🤝 贡献指南

欢迎提交 Issues 和 Pull Requests 来帮助改进项目！

### 贡献流程

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码贡献规范

- 遵循项目的代码风格
- 添加必要的测试用例
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 MIT 许可证 - 详情见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 感谢 [Fiber](https://github.com/gofiber/fiber) 提供的高性能Web框架
- 感谢 [GORM](https://gorm.io/) 提供的优秀ORM库
- 感谢所有开源项目和贡献者的支持

## 📞 联系方式

- 👨‍💻 **开发者**: bobochang
- 📧 **邮箱**: can4hou6joeng4@163.com