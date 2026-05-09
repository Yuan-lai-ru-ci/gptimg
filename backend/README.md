# GPT Image2 Backend

基于Go + Gin的图片生成API后端服务。

## 功能特性

- 用户认证（JWT）
- 图片生成（ChatGPT image2 API）
- 会话管理
- 使用统计
- API配置管理
- 积分系统

## 技术栈

- Go 1.21+
- Gin Web Framework
- SQLite数据库
- JWT认证
- AES-256-GCM加密

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑.env文件，设置JWT密钥和加密密钥
```

**重要**: 生产环境必须修改以下配置：
- `JWT_SECRET`: 至少32字符的随机字符串
- `ENCRYPTION_KEY`: 必须是32字节（32个字符）的随机字符串

### 运行服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API文档

### 认证接口

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/me` - 获取当前用户信息

### 图片生成

- `POST /api/v1/images/generate` - 生成图片
- `GET /api/v1/images/:id` - 获取图片详情
- `DELETE /api/v1/images/:id` - 删除图片

### 会话管理

- `GET /api/v1/sessions` - 获取会话列表
- `POST /api/v1/sessions` - 创建新会话
- `GET /api/v1/sessions/:id` - 获取会话详情
- `GET /api/v1/sessions/:id/messages` - 获取会话消息
- `DELETE /api/v1/sessions/:id` - 删除会话

### 统计数据

- `GET /api/v1/stats/overview` - 统计概览
- `GET /api/v1/stats/daily` - 每日统计

### API配置（管理员）

- `GET /api/v1/config` - 获取配置列表
- `POST /api/v1/config` - 创建配置
- `PUT /api/v1/config/:id` - 更新配置
- `DELETE /api/v1/config/:id` - 删除配置

## 项目结构

```
backend/
├── cmd/server/          # 应用入口
├── internal/
│   ├── api/            # HTTP处理器和路由
│   ├── models/         # 数据模型
│   ├── services/       # 业务逻辑
│   ├── repository/     # 数据访问层
│   ├── database/       # 数据库连接
│   ├── config/         # 配置管理
│   └── utils/          # 工具函数
├── pkg/response/       # 统一响应格式
├── storage/images/     # 图片存储
└── data/              # SQLite数据库
```

## 数据库

使用SQLite作为数据库，首次运行时会自动创建表结构。

## 图片存储

图片默认存储在 `./storage/images` 目录，按用户ID和日期组织：
```
storage/images/{user_id}/{year}/{month}/{filename}
```

## 安全性

- 密码使用bcrypt哈希
- API密钥使用AES-256-GCM加密存储
- JWT token有效期24小时
- 支持CORS配置

## 开发

### 添加新的API接口

1. 在 `internal/api/handlers` 创建handler
2. 在 `internal/api/routes.go` 注册路由
3. 如需数据库操作，在 `internal/repository` 添加方法

### 数据库迁移

数据库迁移在 `internal/database/db.go` 中定义，启动时自动执行。
