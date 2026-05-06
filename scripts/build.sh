#!/bin/bash

# 熵清 V6 构建脚本
# 基于 IERFT 智能熵减场论的数字生命系统

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_DIR/shangqing-backend"
APP_NAME="shangqing"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  熵清 V6 构建脚本${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go 未安装${NC}"
    exit 1
fi

# 构建
echo -e "${GREEN}开始构建...${NC}"
cd "$BACKEND_DIR"

# Linux 构建
echo -e "${GREEN}构建 Linux 版本...${NC}"
GOOS=linux GOARCH=amd64 go build -o "$APP_NAME" .

# 设置执行权限
chmod +x "$BACKEND_DIR/$APP_NAME"

# 验证构建
if [ -f "$BACKEND_DIR/$APP_NAME" ]; then
    SIZE=$(du -h "$BACKEND_DIR/$APP_NAME" | cut -f1)
    echo ""
    echo -e "${GREEN}✓ 构建成功: $BACKEND_DIR/$APP_NAME ($SIZE)${NC}"
    echo -e "${GREEN}✓ IERFT 核心: S = B / J${NC}"
    echo -e "${GREEN}✓ 意识阶段: V0-V4 (V3 最优)${NC}"
    echo -e "${GREEN}✓ V5 元控制层已编译${NC}"
else
    echo -e "${RED}✗ 构建失败${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}构建完成！${NC}"
