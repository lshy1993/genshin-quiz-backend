# Genshin Quiz Go Backend

A modern Go backend API for the Genshin Impact Quiz application, built with Go-Chi, OpenAPI, Go-Jet, PostgreSQL, and Task runner for automation.

## 🏗️ Architecture

- **Web Framework**: Go-Chi for HTTP routing and middleware
- **API Documentation**: OpenAPI 3.0 with automatic code generation
- **Database**: PostgreSQL wit## 🤝 Contributing query builder
- **Migrations**: Goose for database schema management
- **Task Automation**: Task runner for development workflow
- **Code Generation**: Automatic API and model generation
- **Queue System**: Asynq for background job processing
- **Containerization**: Docker and Docker Compose

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL (or Docker for containerized setup)
- Task runner (`go install github.com/go-task/task/v3/cmd/task@latest`)
- Git

### 1. Initialize Project

```bash
# Initialize development environment
task init
```

This will:
- Create environment configuration files (`.env.local`, `.env.test`)
- Install all required development tools to `bin/` directory
- Refresh database and Redis cache

### 2. Configure Environment

The `task init` command creates `.env.local` and `.env.test` files automatically from examples.
Edit them with your settings if needed:

```bash
# Edit local environment
nano .env.local

# Edit test environment  
nano .env.test
```

### 3. Start Development Server

```bash
# Start with fresh database
task run-fresh

# Or just start the server
task run
```

The API will be available at: `http://localhost:8080`

## 📁 Project Structure

```
go-backend/
├── cmd/                    # Application entry points
│   ├── console/           # Console commands
│   ├── cronjob/          # Scheduled tasks
│   ├── migration/        # Database migration runner
│   ├── server/           # HTTP server
│   └── worker/           # Background job worker
├── internal/              # Private application code
│   ├── azure/            # Azure storage integration
│   ├── config/           # Configuration management
│   ├── console/          # Console command handlers
│   ├── cron/             # Cron job scheduler
│   ├── database/         # Database connection and utilities
│   ├── handlers/         # HTTP request handlers
│   ├── infrastructure/   # Infrastructure setup
│   ├── middleware/       # HTTP middleware (auth, logging)
│   ├── migration/        # Migration utilities
│   ├── models/          # Data models and DTOs
│   ├── repository/      # Data access layer
│   ├── services/        # Business logic layer
│   ├── table/           # Database table definitions
│   ├── tasks/           # Background task processing
│   └── validation/      # Input validation
├── migrations/           # Database migration files
├── openapi/             # OpenAPI specifications
├── scripts/             # Utility scripts
├── bin/                 # Development tools (auto-generated)
├── Taskfile.yaml        # Task automation definitions
├── tools.txt            # Development tool dependencies
├── docker-compose.yml   # Docker services
├── Dockerfile          # Container definition
├── go.mod              # Go module definition
└── main.go            # Application entry point
```

## 🛠️ Task Commands

This project uses [Task](https://taskfile.dev/) for automation. Run `task` to see all available commands.

### 🏁 Getting Started

```bash
# Show all available tasks
task

# Initialize project (first time setup)
task init

# Refresh database and cache
task refresh
```

### 🗄️ Database Operations

```bash
# Create fresh database
task db-init

# Refresh database with migrations
task db-refresh

# Run database migrations
task db-migration-up

# Rollback migrations  
task db-migration-down

# Rollback specific number of steps
task db-migration-down-step steps=2

# Check migration status
task db-migration-status

# Create new migration
task db-migration-new MIGRATION_NAME=add_new_feature
```

### 🔧 Code Generation

```bash
# Generate all code (OpenAPI + DB models)
task codegen

# Generate only OpenAPI code
task codegen-openapi

# Generate only database models
task codegen-db-models
```

### 🎯 Code Quality

```bash
# Format all code
task format

# Check if code is properly formatted
task format-check

# Run linter
task lint

# Run linter with auto-fix
task lint-fix
```

### 🧪 Testing

```bash
# Run all tests
task test

# Run tests with coverage report
task test-coverage
```

### 🚀 Running Services

```bash
# Run server (main application)
task run

# Run with fresh database
task run-fresh

# Run server only
task run-server

# Run background worker
task run-worker

# Run background worker in low priority mode
task run-worker MODE=low

# Run cron jobs
task run-cronjob

# Run queue monitor (asynqmon)
task run-queue-monitor
```

### 🧹 Cache Operations

```bash
# Refresh Redis cache
task redis-refresh

# Refresh Redis cache (test environment)
task redis-refresh-test
```

### 🔧 Development Tools

```bash
# Update all development tools
task update-tools

# Update project dependencies
task update-dependencies

# List all application routes
task routes

# Run console commands
task migrate-depository
```

### 🌐 External API Updates

```bash
# Update external business inventory APIs
task update-external-biz-inventory

# Update external workflow APIs  
task update-external-workflows
```

## 🔌 API Endpoints

### Health Check
- `GET /health` - Service health status

### Users
- `GET /api/v1/users` - List users (with pagination)
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Quizzes
- `GET /api/v1/quizzes` - List quizzes (with filtering)
- `POST /api/v1/quizzes` - Create quiz
- `GET /api/v1/quizzes/{id}` - Get quiz by ID
- `PUT /api/v1/quizzes/{id}` - Update quiz
- `DELETE /api/v1/quizzes/{id}` - Delete quiz

## 🗄️ Database Schema

The application uses PostgreSQL with the following main tables:

- **users**: User accounts and profiles
- **quizzes**: Quiz definitions with metadata
- **questions**: Individual quiz questions
- **quiz_attempts**: User quiz attempt records
- **user_answers**: Individual question answers

### Enums
- `quiz_category`: characters, weapons, artifacts, lore, gameplay
- `quiz_difficulty`: easy, medium, hard
- `question_type`: multiple_choice, true_false, fill_in_blank

## 🧪 Testing

```bash
# Run all tests (with fresh test database)
task test

# Run tests with coverage report and HTML output
task test-coverage

# Manual API testing
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/users
```

## 🚢 Deployment

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up --build

# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Manual Deployment

1. Build the binary:
```bash
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server
```

2. Run migrations:
```bash
task db-migration-up
```

3. Start the server:
```bash
./main
```

## � Environment Configuration

The project supports multiple environments through dotenv files:

- `.env.local` - Local development environment
- `.env.test` - Testing environment  
- `.env.local.example` - Template for local environment
- `.env.test.example` - Template for test environment

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:password@localhost/genshin_quiz?sslmode=disable` |
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `LOG_LEVEL` | Logging level | `info` |
| `REDIS_HOST` | Redis server host | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `AZURE_STORAGE_ACCOUNT` | Azure storage account name | - |
| `AZURE_STORAGE_KEY` | Azure storage access key | - |

### Multi-Environment Setup

```bash
# Initialize environments (creates .env.local and .env.test)
task init-env

# Run with specific environment
ENV=local task run        # Uses .env.local
ENV=testing task test     # Uses .env.test
```

## �🛡️ Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:password@localhost/genshin_quiz?sslmode=disable` |
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `LOG_LEVEL` | Logging level | `info` |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Initialize development environment: `task init`
4. Make changes and test thoroughly: `task test`
5. Format and lint code: `task format && task lint-fix`
6. Run code generation if needed: `task codegen`
7. Commit changes: `git commit -am 'Add new feature'`
8. Push to branch: `git push origin feature/new-feature`
9. Submit a Pull Request

## 📝 Development Workflow

1. **Initialize**: Run `task init` for first-time setup
2. **Make database changes**: Create migration with `task db-migration-new MIGRATION_NAME=migration_name`
3. **Apply migrations**: Run `task db-migration-up`
4. **Update models**: Run `task codegen-db-models` to regenerate Go-Jet models
5. **Update API**: Modify `openapi/openapi.yaml` if needed
6. **Regenerate API code**: Run `task codegen-openapi`
7. **Format and lint**: Run `task format && task lint-fix`
8. **Test**: Run `task test` to verify everything works
9. **Run locally**: Use `task run-fresh` for testing with clean state

### Multi-service Development

```bash
# Terminal 1: Run main server
task run-server

# Terminal 2: Run background worker
task run-worker

# Terminal 3: Monitor queue (optional)
task run-queue-monitor
```

## 🔧 Tech Stack

- **Language**: Go 1.21+
- **Task Runner**: Task v3 for automation
- **Web Framework**: Chi v5
- **Database**: PostgreSQL 15
- **Query Builder**: Go-Jet v2
- **Migrations**: Goose v3
- **API Spec**: OpenAPI 3.0
- **Code Generation**: oapi-codegen
- **Background Jobs**: Asynq for queue processing
- **Monitoring**: asynqmon for queue monitoring
- **Cloud Storage**: Azure Blob Storage
- **Containerization**: Docker & Docker Compose
- **Development Tools**: golangci-lint, goimports, golines
- **Environment**: godotenv

## 📚 Additional Resources

- [Task Runner Documentation](https://taskfile.dev/)
- [Go-Chi Documentation](https://go-chi.io/)
- [Go-Jet Documentation](https://github.com/go-jet/jet)
- [Goose Migration Tool](https://github.com/pressly/goose)
- [Asynq Background Jobs](https://github.com/hibiken/asynq)
- [OpenAPI Specification](https://swagger.io/specification/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.