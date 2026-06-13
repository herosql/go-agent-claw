#!/usr/bin/env bash
# ==============================================================================
# deps-start.sh — 依赖服务启动脚本（本项目无自建服务，此脚本为健康检查）
# 用法: bash scripts/deps-start.sh
# ==============================================================================

set -e

log_info()  { echo "[INFO] $1"; }
log_ok()    { echo "[OK]   $1"; }
log_err()   { echo "[ERR]  $1"; }

log_info "检查外部 API 连通性..."

# 1. 智谱 GLM API 健康检查
if curl -s --connect-timeout 5 \
  "https://open.bigmodel.cn/api/paas/v4/models" \
  -o /dev/null -w "%{http_code}" 2>/dev/null | grep -qE "^[234]"; then
    log_ok "智谱 GLM API 可达"
else
    log_err "智谱 GLM API 不可达（网络或 API Key 问题）"
fi

# 2. 飞书开放平台健康检查
if curl -s --connect-timeout 5 \
  "https://open.feishu.cn" \
  -o /dev/null -w "%{http_code}" 2>/dev/null | grep -qE "^[234]"; then
    log_ok "飞书开放平台 API 可达"
else
    log_err "飞书开放平台 API 不可达（网络问题）"
fi

log_ok "依赖服务检查完成。本项目无自建服务（无 DB/Redis/队列）。"
log_info "如需后台常驻进程，参见 docker-compose.dev.yml"