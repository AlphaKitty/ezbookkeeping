#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "正在停止服务..."

# 停止后端
pkill -f "ezbookkeeping server run" 2>/dev/null && echo "后端已停止" || echo "后端未在运行"

# 停止前端（vite dev server）
pkill -f "vite" 2>/dev/null && echo "前端已停止" || echo "前端未在运行"

# 清理 PID 文件
rm -f .pid.frontend .pid.backend

echo "全部已停止"
