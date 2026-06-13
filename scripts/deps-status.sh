#!/usr/bin/env bash
# ==============================================================================
# deps-status.sh — 依赖服务状态检查脚本
# 用法: bash scripts/deps-status.sh
# ==============================================================================

set -e
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_api() {
    local name="$1"
    local url="$2"
    local code
    code=$(curl -s --connect-timeout 5 "$url" -o /dev/null -w "%{http_code}" 2>/dev/null || echo "000")
    if [[ "$code" =~ ^[234] ]]; then
        echo -e "${GREEN}[UP]   ${name}${NC}  HTTP $code"
    else
        echo -e "${RED}[DOWN] ${name}${NC}  HTTP $code"
    fi
}

echo "=== 外部 API 状态 ==="
check_api "智谱 GLM (OpenAI 协议)" "https://open.bigmodel.cn/api/paas/v4/models"
check_api "智谱 GLM (Claude 协议)" "https://open.bigmodel.cn/api/anthropic/models"
check_api "飞书开放平台"            "https://open.feishu.cn"

echo ""
echo "=== Go 环境 ==="
go version | awk '{print "[OK]   Go 版本:", $3}'

echo ""
echo "=== 项目编译状态 ==="
cd "$(dirname "${BASH_SOURCE[0]}")/.."
if go build ./... 2>/dev/null; then
    echo -e "${GREEN}[OK]   所有包编译成功${NC}"
else
    echo -e "${RED}[ERR]  编译失败${NC}"
fi

echo ""
echo "=== 环境变量 ==="
for var in ZHIPU_API_KEY FEISHU_APP_ID FEISHU_APP_SECRET; do
    val="${!var}"
    if [ -z "$val" ]; then
        echo -e "${YELLOW}[WARN] ${var} 未设置${NC}"
    else
        echo -e "${GREEN}[SET]  ${var}${NC}"
    fi
done