#!/bin/bash
set -e

cd "$(dirname "$0")"

# 创建运行时需要的目录
mkdir -p storage log data

# 启动服务
nohup go run . server run --conf-path conf/ezbookkeeping.dev.ini > app.log 2>&1 &

echo "PID: $!"
echo "日志文件: app.log"
echo "查看日志: tail -f app.log"
