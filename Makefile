.PHONY: build run dev test docker-up docker-down clean lint fmt

# 构建
build:
	go build -ldflags="-s -w" -o bin/shangqing ./cmd/server

# 运行（需要先启动 docker-compose）
run:
	./bin/shangqing -c config/config.yaml

# 开发模式（热编译）
dev:
	air -c .air.toml

# 测试
test:
	go test -v -race ./...

# Docker
docker-up:
	docker-compose up -d
	@echo "Waiting for MySQL and Redis..."
	@sleep 10

docker-down:
	docker-compose down

docker-build:
	docker build -t shangqing:v1.0.0 .

docker-run: docker-up
	@sleep 5
	docker run -d --name shangqing-app --network shangqing-net -p 8080:8080 \
		-e COZE_API_KEY=your_key \
		-v $(PWD)/config:/app/config \
		shangqing:v1.0.0

# 清理
clean:
	rm -rf bin/
	docker-compose down -v
	docker rmi shangqing:v1.0.0 || true

# 代码格式
fmt:
	go fmt ./...

# lint
lint:
	golangci-lint run ./...

# proto 生成（未来）
proto:
	@echo "Proto generation not configured yet"

# 依赖
deps:
	go mod tidy
	go mod download
