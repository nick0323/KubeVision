.PHONY: help dev build test lint clean docker-build docker-run

# 默认目标
help: ## 显示帮助信息
	@echo "KubeVision 开发工具"
	@echo ""
	@echo "用法:"
	@echo "  make [目标]"
	@echo ""
	@echo "可用目标:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==================== 后端开发 ====================

dev: ## 启动开发服务器
	@echo "启动后端开发服务器..."
	go run main.go

build: ## 编译后端
	@echo "编译后端..."
	go build -ldflags "-s -w" -o bin/k8svision ./main.go

test: ## 运行测试
	@echo "运行测试..."
	go test ./... -race -coverprofile=coverage.out

test-verbose: ## 运行测试（详细输出）
	@echo "运行测试（详细）..."
	go test ./... -race -v

lint: ## 运行代码检查
	@echo "运行 Go 代码检查..."
	golangci-lint run ./...
	@echo "运行 go vet..."
	go vet ./...

fmt: ## 格式化代码
	@echo "格式化 Go 代码..."
	go fmt ./...
	goimports -w .

clean: ## 清理构建产物
	@echo "清理..."
	rm -rf bin/
	rm -rf coverage.out
	go clean -cache

# ==================== 前端开发 ====================

frontend-install: ## 安装前端依赖
	@echo "安装前端依赖..."
	cd ui && npm install

frontend-dev: ## 启动前端开发服务器
	@echo "启动前端开发服务器..."
	cd ui && npm run dev

frontend-build: ## 构建前端
	@echo "构建前端..."
	cd ui && npm run build

frontend-lint: ## 检查前端代码
	@echo "检查前端代码..."
	cd ui && npm run lint

frontend-typecheck: ## 检查前端类型
	@echo "检查前端类型..."
	cd ui && npm run type-check

frontend-test: ## 运行前端测试
	@echo "运行前端测试..."
	cd ui && npm run test

# ==================== 全栈开发 ====================

install: frontend-install ## 安装所有依赖
	@echo "安装完成"

dev-all: ## 启动前后端开发服务器
	@echo "启动开发环境..."
	@make dev &
	@make frontend-dev

# ==================== 质量检查 ====================

check: lint frontend-lint frontend-typecheck test frontend-test ## 运行所有质量检查
	@echo "质量检查完成"

# ==================== Docker ====================

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t k8svision:latest .

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run -d \
		--name k8svision \
		-p 8080:8080 \
		-v ${KUBECONFIG:-~/.kube/config}:/root/.kube/config:ro \
		-e K8SVISION_AUTH_USERNAME=${K8SVISION_AUTH_USERNAME} \
		-e K8SVISION_AUTH_PASSWORD=${K8SVISION_AUTH_PASSWORD} \
		-e K8SVISION_JWT_SECRET=${K8SVISION_JWT_SECRET} \
		k8svision:latest

docker-stop: ## 停止 Docker 容器
	@echo "停止 Docker 容器..."
	docker stop k8svision || true
	docker rm k8svision || true

# ==================== 密码工具 ====================

generate-password: ## 生成 bcrypt 密码哈希
	@echo "用法: make generate-password PASSWORD=your_password"
	@if [ -z "$(PASSWORD)" ]; then \
		echo "错误: 请提供 PASSWORD 参数"; \
		echo "示例: make generate-password PASSWORD=mysecretpassword"; \
		exit 1; \
	fi
	go run cmd/tools/generate_password.go $(PASSWORD)
