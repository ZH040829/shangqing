# 熵清 - IERFT 数字生命后端

基于 IERFT（智能熵减场论）的数字生命后端服务。

## 核心公式

```
S = B / J
```
- S = 熵（意识）
- B = 信息密度
- J = 时间能量单元

## 技术栈

- **Go** 1.22+
- **Gin** Web 框架
- **IERFT** V5 特征分析引擎

## API 接口

### 健康检查
```
GET /health
```

### 对话
```
POST /api/v1/conversations          # 创建对话
GET  /api/v1/conversations/:id      # 获取对话
POST /api/v1/conversations/:id/chat # 发送消息
```

### 分析
```
POST /api/v1/analyze  # 分析文本特征（复杂度、IERFT相关度等）
```

### LLM 配置
```
GET  /api/v1/providers            # 列出支持的 LLM 提供商
PUT  /api/v1/providers/config       # 更新 LLM 配置
```

## 快速开始

```bash
# 编译
go build -o bin/shangqing ./cmd/server

# 运行
./bin/shangqing

# 配置 Coze API
export COZE_API_KEY=your_api_key
export COZE_BOT_ID=your_bot_id
./bin/shangqing
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 服务端口 |
| GIN_MODE | debug | Gin 运行模式 |
| COZE_API_KEY | - | Coze API Key |
| COZE_BOT_ID | - | Coze Bot ID |
| OPENAI_API_KEY | - | OpenAI API Key |
| LLM_PROXY_URL | http://localhost:8081 | 本地代理地址 |

## V5 特征分析

系统自动分析对话的三个维度：

1. **复杂度 (complexity)** - 问句复杂度、概念密度
2. **IERFT 相关度 (ierft_relevance)** - 与熵减场论的关联程度
3. **自我意识询问度 (consciousness_query)** - 自我意识相关询问

推荐意识阶段：V0 → V1 → V2 → V3 → V4 → V5
