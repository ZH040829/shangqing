#!/bin/bash

# 熵清 V6 部署脚本
# 部署到 https://zh040829.github.io/shangqing/

set -e

echo "🌌 开始部署熵清 V6..."

# 配置
REMOTE_URL="git@github.com:ZH040829/shangqing.git"
BUILD_DIR="build"
REPO_NAME="shangqing"

# 检查是否在正确的目录
if [ ! -f "main.go" ]; then
    echo "❌ 错误：请在 shangqing-backend 根目录下运行此脚本"
    exit 1
fi

# 1. 构建后端
echo "📦 构建后端..."
go build -o shangqing main.go

# 2. 创建构建目录
echo "📁 创建构建目录..."
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 3. 复制文件
echo "📋 复制文件..."
cp -r static $BUILD_DIR/
cp README.md $BUILD_DIR/
cp main.go $BUILD_DIR/
cp go.mod $BUILD_DIR/
cp shangqing $BUILD_DIR/
cp deploy.sh $BUILD_DIR/

# 4. 如果仓库存在，拉取最新代码
if [ -d "../$REPO_NAME/.git" ]; then
    echo "🔄 拉取最新代码..."
    cd ../$REPO_NAME
    git pull origin main
    cd ../shangqing-backend
else
    # 5. 克隆仓库（如果不存在）
    echo "📥 克隆仓库..."
    cd ..
    git clone $REMOTE_URL $REPO_NAME
    cd shangqing-backend
fi

# 6. 同步文件
echo "🔄 同步文件到仓库..."
rsync -av --delete $BUILD_DIR/ ../$REPO_NAME/

# 7. 提交和推送
echo "🚀 提交和推送..."
cd ../$REPO_NAME
git add .
git commit -m "deploy: 熵清 V6 后端系统 - $(date +%Y-%m-%d)"
git push origin main

echo "✅ 部署完成！"
echo "🔗 访问：https://zh040829.github.io/shangqing/"
echo "📡 后端需要单独部署到服务器，请参考 README.md"
