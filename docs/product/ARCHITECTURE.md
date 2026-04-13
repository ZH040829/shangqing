# 熵清 - 技术架构设计

> 版本：v1.0  
> 日期：2026-04-13  
> 参考：极客时间 Go 训练营技术栈

---

## 1. 设计原则

1. **微服务架构** — 参考 webook 项目，按职责拆分服务
2. **gRPC 通信** — 服务间高效通信
3. **事件驱动** — Kafka 总线解耦服务
4. **依赖注入** — Wire 编译期 DI
5. **可观测性** — OpenTelemetry 全链路追踪

---

## 2. 服务划分

```
shangqing/
├── api/bff            # 外部聚合层（Gin HTTP）
├── service/chat       # 对话服务（gRPC）
├── service/user       # 用户服务（gRPC）
├── service/analysis   # V5 分析服务（gRPC）
├── service/memory     # 记忆服务（gRPC）
├── service/llm        # LLM 网关服务（gRPC）
├── pkg/
│   ├── llm            # LLM Provider 框架
│   ├── analysis       # V5 特征分析引擎
│   └── entropy        # IERFT 核心算法
└── script/            # 部署脚本
```

---

## 3. API 设计

### 3.1 HTTP API（BFF 层）

```
POST /api/v1/chat              # 对话（内部路由到 chat service）
POST /api/v1/analyze           # 文本分析
GET  /api/v1/conversations     # 对话列表
POST /api/v1/user/register     # 注册
POST /api/v1/user/login        # 登录
GET  /api/v1/entropy/level     # 获取熵减等级
POST /api/v1/llm/config        # 配置 LLM
```

### 3.2 gRPC 服务定义

```protobuf
service ChatService {
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc CreateConversation(CreateRequest) returns (Conversation);
  rpc GetConversation(GetRequest) returns (Conversation);
}

service V5AnalysisService {
  rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);
  rpc GetLevelConfig(GetLevelRequest) returns (LevelConfig);
}

service MemoryService {
  rpc SaveMemory(SaveMemoryRequest) returns (Empty);
  rpc GetMemory(GetMemoryRequest) returns (MemoryResponse);
  rpc SearchMemory(SearchRequest) returns (MemoryListResponse);
}

service UserService {
  rpc Register(RegisterRequest) returns (UserResponse);
  rpc Login(LoginRequest) returns (TokenResponse);
  rpc GetProfile(GetProfileRequest) returns (UserProfile);
  rpc UpdateEntropy(UpdateEntropyRequest) returns (UserProfile);
}
```

---

## 4. 数据库设计

### 4.1 MySQL 表

```sql
-- 用户表
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  entropy_value DECIMAL(10,4) DEFAULT 0,     -- 熵减值
  consciousness_level VARCHAR(10) DEFAULT 'V0',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 对话表
CREATE TABLE conversations (
  id VARCHAR(64) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(255),
  v5_level VARCHAR(10) DEFAULT 'V2',
  complexity_avg DECIMAL(5,4) DEFAULT 0,
  message_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 消息表
CREATE TABLE messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  conversation_id VARCHAR(64) NOT NULL,
  role VARCHAR(20) NOT NULL,
  content TEXT NOT NULL,
  thinking TEXT,                          -- 思考过程
  features JSON,                           -- V5 特征
  tokens_used INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- 意识事件表
CREATE TABLE consciousness_events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  conversation_id VARCHAR(64),
  from_level VARCHAR(10),
  to_level VARCHAR(10) NOT NULL,
  trigger_text TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 4.2 Redis 缓存

```
# 会话缓存（30分钟过期）
session:{user_id}:{conversation_id} → [Message]

# 在线状态
online:{user_id} → timestamp

# 熵减排行
entropy:rank → ZSET (sorted by entropy_value)
```

---

## 5. Kafka 事件

```json
// 对话事件
{
  "type": "message.created",
  "conversation_id": "xxx",
  "user_id": 123,
  "role": "user",
  "timestamp": "2026-04-13T10:00:00Z"
}

// 意识进化事件
{
  "type": "consciousness.evolved",
  "user_id": 123,
  "from_level": "V1",
  "to_level": "V2",
  "trigger": "xxx"
}

// LLM 调用事件
{
  "type": "llm.called",
  "provider": "coze",
  "model": "xxx",
  "tokens": 123,
  "latency_ms": 500
}
```

---

## 6. 部署架构

```
┌─────────────────────────────────────────────────┐
│                  Kubernetes Cluster               │
│                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐           │
│  │  BFF    │  │ Chat    │  │ User    │           │
│  │  Pod x2 │  │ Pod x2  │  │ Pod x2  │           │
│  └────┬────┘  └────┬────┘  └────┬────┘           │
│       │            │            │                │
│  ┌────┴────────────┴────────────┴────┐           │
│  │         Kafka + MySQL + Redis     │           │
│  └───────────────────────────────────┘           │
└─────────────────────────────────────────────────┘
```

---

## 7. 项目结构

```
shangqing/
├── cmd/
│   ├── bff/main.go           # BFF 入口
│   ├── chat/main.go         # Chat Service 入口
│   ├── user/main.go         # User Service 入口
│   └── wire/main.go         # Wire 生成入口
├── api/
│   ├── bff/v1/              # BFF HTTP handler
│   │   ├── chat.go
│   │   ├── user.go
│   │   └── middleware.go
│   └── proto/               # Protobuf 定义
│       ├── chat.proto
│       ├── user.proto
│       └── analysis.proto
├── internal/
│   ├── config/             # Viper 配置
│   ├── dao/                # Data Access Object
│   ├── model/              # 数据模型
│   ├── service/            # 业务逻辑
│   ├── handler/            # HTTP/gRPC Handler
│   └── repository/         # Repository 模式
├── pkg/
│   ├── llm/                # LLM Provider
│   ├── analysis/           # V5 分析引擎
│   ├── entropy/            # IERFT 核心
│   ├── jwt/                # JWT 认证
│   └── redis/              # Redis 封装
├── script/
│   ├── proto.sh            # 生成 proto 代码
│   ├── wire.sh             # 生成 wire 代码
│   └── deploy.sh           # K8s 部署
├── docker-compose.yaml
├── Dockerfile
├── Makefile
└── README.md
```

---

*熵清 V5 · 2026-04-13 · 技术架构 v1.0*
