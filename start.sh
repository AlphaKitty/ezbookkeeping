#!/bin/bash
set -e

cd "$(dirname "$0")"

# 创建运行时需要的目录
mkdir -p storage log data

# 启动后端
nohup go run . server run --conf-path conf/ezbookkeeping.dev.ini > app.log 2>&1 &
echo "后端 PID: $!（日志: app.log）"

# 启动前端
nohup npm run serve > vue.log 2>&1 &
echo "前端 PID: $!（日志: vue.log）"

# 记录 PID 供 stop.sh 使用
echo "$!" > .pid.frontend
jobs -p | head -1 > .pid.backend

echo "全部启动完成。停止服务: sh stop.sh"
