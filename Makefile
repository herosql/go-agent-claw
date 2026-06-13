# ==============================================================================
# Makefile — go-agent-claw 构建与开发入口
#
# 用法:
#   make help          # 查看所有可用命令
#   make deps/install  # 安装依赖
#   make deps/status   # 检查外部 API + Go 环境 + 编译状态
#   make deps/start    # 启动外部依赖（健康检查）
#   make deps/stop     # 停止外部依赖（本项目无自建服务，空操作）
#   make smoke         # 冒烟测试
#   make build         # 编译所有包
#   make clean         # 清理构建产物
#
# 环境变量（可选）:
#   ZHIPU_API_KEY      智谱 GLM API Key
#   FEISHU_APP_ID      飞书应用 App ID
#   FEISHU_APP_SECRET  飞书应用密钥
# ==============================================================================

.PHONY: help deps/install deps/status deps/start deps/stop smoke build clean

# 默认目标
help:
	@echo "go-agent-claw — 可用命令"
	@echo ""
	@echo "  make deps/install   安装 Go 依赖并验证编译"
	@echo "  make deps/status    检查外部 API 状态 + Go 环境 + 编译状态"
	@echo "  make deps/start     检查外部 API 连通性（本项目无自建服务）"
	@echo "  make deps/stop      无操作（本项目无自建服务需停止）"
	@echo "  make smoke          冒烟测试：envcheck / 编译 / CLI / feishu / bench"
	@echo "  make build          编译所有包"
	@echo "  make clean          清理构建产物"
	@echo ""
	@echo "  make lint           代码检查（需 golangci-lint）"
	@echo "  make test           运行所有测试"
	@echo ""

# ---- 依赖管理 ----------------------------------------------------------------

deps/install:
	@bash scripts/install-deps.sh

deps/status:
	@bash scripts/deps-status.sh

deps/start:
	@bash scripts/deps-start.sh

deps/stop:
	@bash scripts/deps-stop.sh

smoke:
	@bash scripts/smoke-test.sh

# ---- 编译 --------------------------------------------------------------------

build:
	go build ./...

clean:
	go clean
	rm -f bin/claw bin/feishu bin/agentops bin/bench bin/envcheck 2>/dev/null || true

# ---- 开发辅助 ----------------------------------------------------------------

lint:
	golangci-lint run ./... || echo "[WARN] golangci-lint 未安装，跳过"

test:
	go test ./...

# 交叉编译示例（按需启用）
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/claw-linux ./cmd/claw

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/claw.exe ./cmd/claw