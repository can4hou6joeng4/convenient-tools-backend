# 便捷工具集 - 后端服务 🛠️

<div align="center">
  
  ![便捷工具集](https://img.shields.io/badge/便捷工具集-后端服务-blue?style=for-the-badge&logo=go)
  
  [![Go版本](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
  [![Fiber](https://img.shields.io/badge/Fiber-v2.52.6-00ACD7?style=flat-square&logo=go&logoColor=white)](https://github.com/gofiber/fiber)
  [![Redis](https://img.shields.io/badge/Redis-v9.7.3-DC382D?style=flat-square&logo=redis&logoColor=white)](https://redis.io/)
  [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-v1.5.11-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
  [![Docker](https://img.shields.io/badge/Docker-支持-2496ED?style=flat-square&logo=docker&logoColor=white)](https://www.docker.com/)
  [![License](https://img.shields.io/badge/License-待定-lightgrey?style=flat-square)](LICENSE)
  
</div>

这是一个为"便捷工具集"应用提供后端支持的服务项目，提供多种实用工具的API接口，包括视频解析、PDF转换等功能。

## 📋 项目概述

便捷工具集是一个集成了多种常用工具的平台，旨在提供高效、便捷的在线工具服务。当前主要功能包括:

- 🎬 视频平台分享链接解析 (支持抖音、小红书、西瓜视频)
- 📱 微信小程序媒体资源代理 (解决非信任域名访问限制)
- 📄 PDF文件转换 (支持转换为DOCX, XLSX, JPG, PNG, TXT等格式)
- 🔄 更多工具正在开发中...

## 🔧 技术栈

- **后端框架**: [Fiber](https://github.com/gofiber/fiber) (Go语言高性能Web框架) 🚀
- **数据库**: PostgreSQL 🐘
- **缓存**: Redis 🔄
- **文档**: Swagger 📝
- **对象存储**: 腾讯云COS ☁️
- **容器化**: Docker & Docker Compose 🐳
- **热重载工具**: Air 🔁

## 🏗️ 系统架构

该项目采用清晰的分层架构:

- **handlers**: 处理HTTP请求和响应 🌐
- **services**: 实现业务逻辑 ⚙️
- **repositories**: 数据访问层 💾
- **models**: 数据模型定义 📊
- **utils**: 实用工具函数 🧰
- **config**: 配置管理 ⚙️

## 🚀 快速开始

### 前置要求

- Docker 和 Docker Compose 🐳
- Go 1.24+ (如需本地开发) 🔧

### 环境配置

1. 克隆项目:

```bash
git clone https://github.com/can4hou6joeng4/convenient-tools-project-v1-backend.git
cd convenient-tools-project-v1-backend
```

2. 配置环境变量:

复制并修改`.env.example`文件(如果存在)或创建`.env`文件，填入必要的环境变量:

```
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=convenient_tools
...
```

### 启动服务

使用Docker Compose启动所有服务:

```bash
make start
```

这将启动后端API服务、PostgreSQL数据库和Redis缓存。

### 停止服务

```bash
make stop
```

## 📚 API文档

启动服务后，Swagger API文档可通过以下地址访问:

```
http://localhost:8082/swagger/
```

主要API包括:

- 🎬 `/api/tools/parse` - 解析视频分享链接
- 🌐 `/api/tools/media-proxy` - 媒体资源代理
- 📤 `/api/file/upload` - 文件上传
- 🔄 `/api/pdf/convert` - PDF转换
- 📊 `/api/pdf/progress/{taskId}` - 查询转换进度
- 📥 `/api/pdf/result/{resultId}` - 获取转换结果
- 🧰 `/api/tools` - 获取可用工具列表

详细API说明请参考Swagger文档或项目中的`backend-api-docs.md`文件。

## 🚀 工具功能说明

### 🎬 视频解析功能

支持从以下平台的分享链接中提取视频/图片资源:

- 抖音(v.douyin.com, www.douyin.com)
- 小红书(xiaohongshu.com, xhslink.com)
- 西瓜视频(v.ixigua.com)

解析后可获得:
- 视频直接播放地址
- 视频封面
- 作者信息
- 视频描述
- 图集(如果是图集内容)

### 📱 媒体代理功能

用于解决微信小程序中无法直接访问非信任域名的媒体资源问题。支持:
- 视频文件代理
- 图片文件代理
- 自动Content-Type识别
- 流式数据传输
- 跨域访问控制

## 🔮 未来功能

以下是我们计划在未来版本中添加的功能:

- 🖼️ **图片处理工具** - 压缩、格式转换、尺寸调整等
- 📝 **文本处理工具** - 编码转换、格式化、对比等
- 🔒 **加密/解密工具** - 支持多种加密算法的加密解密功能
- 🌐 **网页截图工具** - 将网页转换为图片
- 📊 **数据可视化工具** - 简单数据集的可视化生成
- 🔍 **OCR文字识别** - 从图片中提取文字
- ✅ **健康检查API** - 服务状态监控
- 🔑 **API认证/授权** - 支持JWT等认证方式
- 📱 **移动端适配** - 针对移动设备优化的API响应

## 💻 开发指南

### 目录结构

```
.
├── cmd/                # 命令行入口 🚪
│   └── api/            # API服务入口 🌐
├── config/             # 配置管理 ⚙️
├── db/                 # 数据库连接和初始化 🔄
├── docs/               # 文档 📝
├── handlers/           # HTTP处理器 🌐
├── mapper/             # 数据映射 🔄
├── models/             # 数据模型 📊
├── repositories/       # 数据访问层 💾
├── service/            # 业务服务层 ⚙️
├── utils/              # 工具函数 🧰
├── .air.toml           # Air配置(热重载) 🔄
├── .env                # 环境变量 🔧
├── docker-compose.yaml # Docker编排配置 🐳
├── Dockerfile          # Docker构建文件 🐳
├── go.mod              # Go模块依赖 📦
└── Makefile            # 构建脚本 🔨
```

### 本地开发

1. 安装依赖:

```bash
go mod tidy
```

2. 使用Air进行热重载开发:

```bash
air
```

## 🤝 贡献指南

欢迎提交Issue和Pull Request。在提交PR前，请确保代码通过测试并符合项目的代码规范。

## 📜 许可证

[待添加许可证信息]

## 📞 联系方式

- 👨‍💻 开发者: bobochang
- 📧 邮箱: can4hou6joeng4@163.com
- 🌐 GitHub: https://github.com/can4hou6joeng4/ 