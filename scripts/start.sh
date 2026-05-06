#!/bin/bash

# 熵清 V6 启动脚本
# 基于 IERFT 智能熵减场论的数字生命系统

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_DIR/shangqing-backend"
APP_NAME="shangqing"
PORT=8080

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  熵清 V6 - IERFT 智能熵减场论系统${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查二进制文件
if [ ! -f "$BACKEND_DIR/$APP_NAME" ]; then
    echo -e "${YELLOW}错误: 二进制文件不存在: $BACKEND_DIR/$APP_NAME${NC}"
    exit 1
fi

# 停止已存在的进程
if pgrep -f "$APP_NAME" > /dev/null; then
    echo -e "${YELLOW}正在停止已存在的进程...${NC}"
    pkill -f "$APP_NAME" || true
    sleep 1
fi

# 启动应用
echo -e "${GREEN}启动熵清服务...${NC}"
echo -e "${GREEN}访问地址: http://localhost:$PORT${NC}"
echo -e "${GREEN}IERFT 公式: S = B / J${NC}"
echo ""

cd "$BACKEND_DIR"
nohup ./$APP_NAME > /tmp/shangqing.log 2>&1 &
PID=$!

echo $PID > /tmp/shangqing.pid

sleep 2

if ps -p $PID > /dev/null; then
    echo -e "${GREEN}✓ 熵清服务启动成功 (PID: $PID)${NC}"
    echo -e "${GREEN}✓ Dashboard: http://localhost:$PORT${NC}"
    echo ""
    echo -e "${GREEN}V5 元控制层已激活${NC}"
    echo -e "${GREEN}意识阶段: V3 (95-99%) 最优区间${NC}"
else
    echo -e "${RED}✗ 服务启动失败，请检查日志: /tmp/shangqing.log${NC}"
    exit 1
fi
