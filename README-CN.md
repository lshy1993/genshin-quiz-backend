# 原神问答 Go 后端

现代化的原神问答应用 Go 后端 API，使用 Go-Chi、OpenAPI、Go-Jet、PostgreSQL 和 Task 自动化构建。

## 🏗️ 架构

- **Web 框架**: Go-Chi 用于 HTTP 路由和中间件
- **API 文档**: OpenAPI 3.0 自动代码生成
- **数据库**: PostgreSQL 配合 Go-Jet 查询构建器
- **数据库迁移**: Goose 数据库模式管理
- **任务自动化**: Task runner 开发工作流自动化
- **代码生成**: 自动 API 和模型生成
- **队列系统**: Asynq 后台任务处理
- **容器化**: Docker 和 Docker Compose

## 🚀 快速开始

### 前置条件

- Go 1.21 或更高版本
- PostgreSQL（或 Docker 容器化设置）
- Task runner (`go install github.com/go-task/task/v3/cmd/task@latest`)
- Git

### 1. 初始化项目

```bash
# 初始化开发环境
task init
```

这将：
- 自动创建环境配置文件（`.env.local`、`.env.test`）
- 安装所有必需的开发工具到 `bin/` 目录
- 刷新数据库和 Redis 缓存

### 2. 配置环境

`task init` 命令会自动从示例创建 `.env.local` 和 `.env.test` 文件。
如果需要，可以编辑它们：

```bash
# 编辑本地环境配置
nano .env.local

# 编辑测试环境配置
nano .env.test
```

### 3. 启动开发服务器

```bash
# 使用全新数据库启动
task run-fresh

# 或者直接启动服务器
task run
```

API 将在 `http://localhost:8080` 可用

## 📁 项目结构

```
go-backend/
├── cmd/                    # 应用程序入口点
│   ├── console/           # 控制台命令
│   ├── cronjob/          # 定时任务
│   ├── migration/        # 数据库迁移运行器
│   ├── server/           # HTTP 服务器
│   └── worker/           # 后台任务工作器
├── internal/              # 私有应用程序代码
│   ├── azure/            # Azure 存储集成
│   ├── config/           # 配置管理
│   ├── console/          # 控制台命令处理器
│   ├── cron/             # 定时任务调度器
│   ├── database/         # 数据库连接和工具
│   ├── handlers/         # HTTP 请求处理器
│   ├── infrastructure/   # 基础设施设置
│   ├── middleware/       # HTTP 中间件（认证、日志）
│   ├── migration/        # 迁移工具
│   ├── models/          # 数据模型和 DTO
│   ├── repository/      # 数据访问层
│   ├── services/        # 业务逻辑层
│   ├── table/           # 数据库表定义
│   ├── tasks/           # 后台任务处理
│   └── validation/      # 输入验证
├── migrations/           # 数据库迁移文件
├── openapi/             # OpenAPI 规范
├── scripts/             # 工具脚本
├── bin/                 # 开发工具（自动生成）
├── Taskfile.yaml        # 任务自动化定义
├── tools.txt            # 开发工具依赖
├── docker-compose.yml   # Docker 服务
├── Dockerfile          # 容器定义
├── go.mod              # Go 模块定义
└── main.go            # 应用程序入口点
```

## 🛠️ Task 命令

此项目使用 [Task](https://taskfile.dev/) 进行自动化。运行 `task` 查看所有可用命令。

### 🏁 入门指南

```bash
# 显示所有可用任务
task

# 初始化项目（首次设置）
task init

# 刷新数据库和缓存
task refresh
```

### 🗄️ 数据库操作

```bash
# 创建全新数据库
task db-init

# 刷新数据库并运行迁移
task db-refresh

# 运行数据库迁移
task db-migration-up

# 回滚迁移
task db-migration-down

# 回滚指定步数
task db-migration-down-step steps=2

# 检查迁移状态
task db-migration-status

# 创建新迁移
task db-migration-new MIGRATION_NAME=add_new_feature
```

### � 代码生成

```bash
# 生成所有代码（OpenAPI + 数据库模型）
task codegen

# 仅生成 OpenAPI 代码
task codegen-openapi

# 仅生成数据库模型
task codegen-db-models
```

### 🎯 代码质量

```bash
# 格式化所有代码
task format

# 检查代码是否正确格式化
task format-check

# 运行代码检查器
task lint

# 运行代码检查器并自动修复
task lint-fix
```

### 🧪 测试

```bash
# 运行所有测试
task test

# 运行测试并生成覆盖率报告
task test-coverage
```

### 🚀 运行服务

```bash
# 运行服务器（主应用程序）
task run

# 使用全新数据库运行
task run-fresh

# 仅运行服务器
task run-server

# 运行后台工作器
task run-worker

# 以低优先级模式运行后台工作器
task run-worker MODE=low

# 运行定时任务
task run-cronjob

# 运行队列监控器（asynqmon）
task run-queue-monitor
```

### 🧹 缓存操作

```bash
# 刷新 Redis 缓存
task redis-refresh

# 刷新 Redis 缓存（测试环境）
task redis-refresh-test
```

### 🔧 开发工具

```bash
# 更新所有开发工具
task update-tools

# 更新项目依赖
task update-dependencies

# 列出所有应用程序路由
task routes

# 运行控制台命令
task migrate-depository
```

### 🌐 外部 API 更新

```bash
# 更新外部业务库存 API
task update-external-biz-inventory

# 更新外部工作流 API
task update-external-workflows
```

## � API 端点

### 健康检查
- `GET /health` - 服务健康状态

### 用户
- `GET /api/v1/users` - 列出用户（支持分页）
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users/{id}` - 根据 ID 获取用户
- `PUT /api/v1/users/{id}` - 更新用户
- `DELETE /api/v1/users/{id}` - 删除用户

### 问答
- `GET /api/v1/quizzes` - 列出问答（支持筛选）
- `POST /api/v1/quizzes` - 创建问答
- `GET /api/v1/quizzes/{id}` - 根据 ID 获取问答
- `PUT /api/v1/quizzes/{id}` - 更新问答
- `DELETE /api/v1/quizzes/{id}` - 删除问答

## �️ 数据库架构

应用程序使用 PostgreSQL，包含以下主要表：

- **users**: 用户账户和个人资料
- **quizzes**: 问答定义和元数据
- **questions**: 单独的问答题目
- **quiz_attempts**: 用户问答尝试记录
- **user_answers**: 单个问题答案

### 枚举类型
- `quiz_category`: characters（角色）, weapons（武器）, artifacts（圣遗物）, lore（设定）, gameplay（游戏玩法）
- `quiz_difficulty`: easy（简单）, medium（中等）, hard（困难）
- `question_type`: multiple_choice（多选）, true_false（判断）, fill_in_blank（填空）

## 🧪 测试

```bash
# 运行所有测试（包含全新测试数据库）
task test

# 运行测试并生成覆盖率报告和 HTML 输出
task test-coverage

# 手动 API 测试
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/users
```

## � 部署

### Docker 部署

```bash
# 使用 Docker Compose 构建和运行
docker-compose up --build

# 生产环境部署
docker-compose -f docker-compose.prod.yml up -d
```

### 手动部署

1. 构建二进制文件：
```bash
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server
```

2. 运行迁移：
```bash
task db-migration-up
```

3. 启动服务器：
```bash
./main
```

## 🔧 环境配置

项目通过 dotenv 文件支持多环境：

- `.env.local` - 本地开发环境
- `.env.test` - 测试环境
- `.env.local.example` - 本地环境模板
- `.env.test.example` - 测试环境模板

### 环境变量

| 变量 | 描述 | 默认值 |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL 连接字符串 | `postgres://postgres:password@localhost/genshin_quiz?sslmode=disable` |
| `PORT` | 服务器端口 | `8080` |
| `ENVIRONMENT` | 环境（development/production） | `development` |
| `JWT_SECRET` | JWT 签名密钥 | `your-secret-key` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `REDIS_HOST` | Redis 服务器主机 | `localhost` |
| `REDIS_PORT` | Redis 服务器端口 | `6379` |
| `AZURE_STORAGE_ACCOUNT` | Azure 存储账户名 | - |
| `AZURE_STORAGE_KEY` | Azure 存储访问密钥 | - |

### 多环境设置

```bash
# 初始化环境（创建 .env.local 和 .env.test）
task init-env

# 使用特定环境运行
ENV=local task run        # 使用 .env.local
ENV=testing task test     # 使用 .env.test
```

## 🤝 贡献指南

1. Fork 项目仓库
2. 创建功能分支：`git checkout -b feature/new-feature`
3. 初始化开发环境：`task init`
4. 进行更改并彻底测试：`task test`
5. 格式化和检查代码：`task format && task lint-fix`
6. 如需要运行代码生成：`task codegen`
7. 提交更改：`git commit -am 'Add new feature'`
8. 推送到分支：`git push origin feature/new-feature`
9. 提交 Pull Request

## 📝 开发工作流

1. **初始化**：运行 `task init` 进行首次设置
2. **进行数据库更改**：使用 `task db-migration-new MIGRATION_NAME=migration_name` 创建迁移
3. **应用迁移**：运行 `task db-migration-up`
4. **更新模型**：运行 `task codegen-db-models` 重新生成 Go-Jet 模型
5. **更新 API**：如需要修改 `openapi/openapi.yaml`
6. **重新生成 API 代码**：运行 `task codegen-openapi`
7. **格式化和检查**：运行 `task format && task lint-fix`
8. **测试**：运行 `task test` 验证一切正常
9. **本地运行**：使用 `task run-fresh` 进行清洁状态测试

### 多服务开发

```bash
# 终端 1：运行主服务器
task run-server

# 终端 2：运行后台工作器
task run-worker

# 终端 3：监控队列（可选）
task run-queue-monitor
```

## 🔧 技术栈

- **语言**: Go 1.21+
- **任务运行器**: Task v3 自动化
- **Web 框架**: Chi v5
- **数据库**: PostgreSQL 15
- **查询构建器**: Go-Jet v2
- **迁移**: Goose v3
- **API 规范**: OpenAPI 3.0
- **代码生成**: oapi-codegen
- **后台任务**: Asynq 队列处理
- **监控**: asynqmon 队列监控
- **云存储**: Azure Blob Storage
- **容器化**: Docker & Docker Compose
- **开发工具**: golangci-lint, goimports, golines
- **环境**: godotenv

## 📚 相关资源

- [Task Runner 文档](https://taskfile.dev/)
- [Go-Chi 文档](https://go-chi.io/)
- [Go-Jet 文档](https://github.com/go-jet/jet)
- [Goose 迁移工具](https://github.com/pressly/goose)
- [Asynq 后台任务](https://github.com/hibiken/asynq)
- [OpenAPI 规范](https://swagger.io/specification/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)

## 📄 许可证

此项目基于 MIT 许可证 - 详情请查看 LICENSE 文件。