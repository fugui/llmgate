.PHONY: build run test clean docker-up docker-down migrate lint fmt deps

# 变量
BINARY_NAME=llmgate
DOCKER_COMPOSE=docker-compose

# 构建
build:
	go build -o $(BINARY_NAME) cmd/server/main.go

# 运行开发服务器
run:
	go run cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 清理
clean:
	rm -f $(BINARY_NAME)
	go clean

# Docker 命令
docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f server

docker-build:
	$(DOCKER_COMPOSE) build

# 数据库迁移（手动执行 SQL）
migrate:
	psql -h localhost -U llmgate -d llmgate -f migrations/001_init.sql

# 代码检查
lint:
	golangci-lint run

# 格式化代码
fmt:
	go fmt ./...

# 安装依赖
deps:
	go mod download
	go mod tidy
