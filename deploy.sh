#!/bin/bash
set -e

# ============================================================
# Kimi Config Platform - Render 自动部署脚本
# ============================================================
# 使用方法:
#   1. 在 Render Dashboard 获取 API Key: https://dashboard.render.com/u/settings#api-keys
#   2. 设置环境变量: export RENDER_API_KEY=rkp_xxxx
#   3. 运行: ./deploy.sh
# ============================================================

REPO_URL="https://github.com/dormitory219/kimi-config-platform"
BRANCH="main"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 API Key
if [ -z "$RENDER_API_KEY" ]; then
    echo -e "${RED}错误: 请设置 RENDER_API_KEY 环境变量${NC}"
    echo "获取方式: https://dashboard.render.com/u/settings#api-keys"
    echo "然后运行: export RENDER_API_KEY=rkp_xxxx"
    exit 1
fi

echo -e "${GREEN}🚀 开始部署 Kimi Config Platform 到 Render...${NC}"

# ============================================================
# 步骤 1: 部署后端 API 服务
# ============================================================
echo -e "\n${YELLOW}📦 步骤 1/3: 部署后端 API 服务...${NC}"

BACKEND_RESPONSE=$(curl -s -X POST \
  https://api.render.com/v1/services \
  -H "Authorization: Bearer $RENDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"web_service\",
    \"name\": \"kimi-config-api\",
    \"ownerId\": \"usr-default\",
    \"repo\": \"$REPO_URL\",
    \"branch\": \"$BRANCH\",
    \"rootDir\": \"kimi-config-server\",
    \"runtime\": \"docker\",
    \"dockerfilePath\": \"./Dockerfile\",
    \"envVars\": [
      {\"key\": \"PORT\", \"value\": \"8080\"},
      {\"key\": \"CONFIG_REPO_PATH\", \"value\": \"./config-repo\"},
      {\"key\": \"GIN_MODE\", \"value\": \"release\"}
    ]
  }" 2>/dev/null || echo "{}")

# 检查后端是否创建成功
if echo "$BACKEND_RESPONSE" | grep -q '"id"'; then
    BACKEND_ID=$(echo "$BACKEND_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    BACKEND_URL=$(echo "$BACKEND_RESPONSE" | grep -o '"url":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}✅ 后端服务创建成功!${NC}"
    echo "   Service ID: $BACKEND_ID"
    echo "   URL: $BACKEND_URL"
else
    echo -e "${RED}❌ 后端服务创建失败${NC}"
    echo "响应: $BACKEND_RESPONSE"
    exit 1
fi

# ============================================================
# 步骤 2: 等待后端部署完成并获取 URL
# ============================================================
echo -e "\n${YELLOW}⏳ 步骤 2/3: 等待后端部署完成...${NC}"
echo "   这可能需要 2-5 分钟..."

for i in {1..30}; do
    sleep 10
    STATUS_RESPONSE=$(curl -s \
      "https://api.render.com/v1/services/$BACKEND_ID" \
      -H "Authorization: Bearer $RENDER_API_KEY" 2>/dev/null || echo "{}")

    SERVICE_STATUS=$(echo "$STATUS_RESPONSE" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
    SERVICE_URL=$(echo "$STATUS_RESPONSE" | grep -o '"serviceDetails":{[^}]*"url":"[^"]*"' | grep -o '"url":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$SERVICE_URL" ] && [ "$SERVICE_URL" != "null" ]; then
        echo -e "${GREEN}✅ 后端部署完成!${NC}"
        echo "   访问地址: https://$SERVICE_URL"
        BACKEND_DOMAIN="https://$SERVICE_URL"
        break
    fi

    echo "   状态: ${SERVICE_STATUS:-deploying}... (${i}/30)"
done

if [ -z "$BACKEND_DOMAIN" ]; then
    echo -e "${YELLOW}⚠️ 等待超时，但服务可能仍在部署中${NC}"
    echo "   请稍后手动在 Dashboard 查看: https://dashboard.render.com"
    BACKEND_DOMAIN="https://kimi-config-api.onrender.com"
fi

# ============================================================
# 步骤 3: 部署前端静态站点
# ============================================================
echo -e "\n${YELLOW}📦 步骤 3/3: 部署前端静态站点...${NC}"
echo "   API 地址: $BACKEND_DOMAIN"

FRONTEND_RESPONSE=$(curl -s -X POST \
  https://api.render.com/v1/services \
  -H "Authorization: Bearer $RENDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"static_site\",
    \"name\": \"kimi-config-web\",
    \"ownerId\": \"usr-default\",
    \"repo\": \"$REPO_URL\",
    \"branch\": \"$BRANCH\",
    \"rootDir\": \"kimi-config-web\",
    \"buildCommand\": \"npm install && npm run build\",
    \"staticPublishPath\": \"./dist\",
    \"envVars\": [
      {\"key\": \"VITE_API_URL\", \"value\": \"$BACKEND_DOMAIN\"}
    ]
  }" 2>/dev/null || echo "{}")

if echo "$FRONTEND_RESPONSE" | grep -q '"id"'; then
    FRONTEND_ID=$(echo "$FRONTEND_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}✅ 前端服务创建成功!${NC}"
    echo "   Service ID: $FRONTEND_ID"
else
    echo -e "${RED}❌ 前端服务创建失败${NC}"
    echo "响应: $FRONTEND_RESPONSE"
    exit 1
fi

# ============================================================
# 完成
# ============================================================
echo -e "\n${GREEN}🎉 部署完成!${NC}"
echo "=========================================="
echo "  后端 API: $BACKEND_DOMAIN"
echo "  前端页面: https://kimi-config-web.onrender.com"
echo "  Dashboard: https://dashboard.render.com"
echo "=========================================="
echo ""
echo -e "${YELLOW}注意:${NC}"
echo "  - 首次部署可能需要 3-5 分钟完成构建"
echo "  - Render 免费实例 15 分钟无请求会休眠，首次访问可能较慢"
echo "  - 配置数据存储在容器内，重新部署后会重置"
