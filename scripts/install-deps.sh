#!/usr/bin/env bash
# ==============================================================================
# install-deps.sh — Go-Agent-Claw 依赖安装脚本
# 用法: bash scripts/install-deps.sh
# 自主修复原则：同一错误连续出现 3 次才终止执行
# ==============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

MAX_RETRIES=3
RETRIES=0
last_err=""

log_info()  { echo "[INFO]  $1"; }
log_warn()  { echo "[WARN]  $1"; }
log_err()   { echo "[ERR]   $1"; }
log_ok()    { echo "[OK]    $1"; }

run_with_retry() {
    local cmd="$1"
    local description="$2"
    RETRIES=0
    last_err=""

    while true; do
        log_info "正在: $description"
        if eval "$cmd" 2>&1; then
            log_ok "$description 成功"
            return 0
        else
            ((RETRIES++))
            last_err=$(eval "$cmd" 2>&1 | tail -5)
            if [ $RETRIES -ge $MAX_RETRIES ]; then
                log_err "$description 失败（已重试 $MAX_RETRIES 次）: $last_err"
                return 1
            fi
            log_warn "$description 失败（第 ${RETRIES}次），10秒后重试...（最多重试 $MAX_RETRIES 次）"
            sleep 10
        fi
    done
}

# ---- 检查 Go -----------------------------------------------------------
log_info "检查 Go 环境..."
if ! command -v go &> /dev/null; then
    log_err "Go 未安装或不在 PATH 中"
    exit 1
fi
GO_VERSION=$(go version | grep -oP 'go\d+\.\d+' | head -1)
GO_NUMERIC=$(echo "$GO_VERSION" | grep -oP '\d+\.\d+' | head -1)
REQUIRED=1.26
if (( $(echo "$GO_NUMERIC $REQUIRED" | awk '{print ($1 >= $2)}') )); then
    log_ok "Go 版本: $(go version)"
else
    log_err "Go 版本过低，需要 ≥ 1.26，当前: $(go version)"
    exit 1
fi

# ---- 检查 GOPROXY ------------------------------------------------------
log_info "检查 GOPROXY..."
CURRENT_PROXY=$(go env GOPROXY)
if [[ "$CURRENT_PROXY" == *"goproxy.cn"* ]] || [[ "$CURRENT_PROXY" == *"goproxy.io"* ]]; then
    log_ok "GOPROXY 已配置: $CURRENT_PROXY"
else
    log_warn "GOPROXY 未配置国内镜像，设置为 goproxy.cn"
    go env -w GOPROXY=https://goproxy.cn,direct
fi

# ---- go mod download（最多重试 3 次）------------------------------------
run_with_retry "go mod download" "下载 Go 依赖" || {
    log_err "go mod download 失败，尝试备选方案..."
    # 备选：逐个模块下载（针对网络不稳定情况）
    go mod tidy
}

# ---- go mod tidy --------------------------------------------------------
run_with_retry "go mod tidy" "整理 go.mod / go.sum" || exit 1

# ---- go build ./... ----------------------------------------------------
log_info "编译所有包..."
if go build ./... 2>&1; then
    log_ok "全量编译成功"
else
    log_err "全量编译失败，请检查 import 路径"
    exit 1
fi

# ---- 二进制验证 ---------------------------------------------------------
for bin in claw feishu agentops bench envcheck; do
    bin_path="$PROJECT_ROOT/bin/$bin"
    if [ -f "$bin_path" ]; then
        log_ok "bin/$bin 存在"
    else
        log_warn "bin/$bin 未生成（可能 cmd/$bin 不存在或编译未输出二进制）"
    fi
done

log_ok "=============================================="
log_ok "依赖安装完成！"
log_ok "下一步: bash scripts/deps-start.sh（可选，如需外部服务）"
log_ok "=============================================="