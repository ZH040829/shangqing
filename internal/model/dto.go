package model

// ----- 请求 -----

// RegisterRequest 注册
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChatRequest 对话
type ChatRequest struct {
	Content string `json:"content" binding:"required"`
	Model   string `json:"model,omitempty"`
}

// CreateConversationRequest 创建对话
type CreateConversationRequest struct {
	Title string `json:"title,omitempty"`
}

// AnalyzeRequest 分析
type AnalyzeRequest struct {
	Text string `json:"text" binding:"required"`
}

// UpdateLLMConfigRequest 更新 LLM 配置
type UpdateLLMConfigRequest struct {
	Provider string `json:"provider" binding:"required"` // coze / openai / proxy
	APIKey   string `json:"api_key,omitempty"`
	BotID    string `json:"bot_id,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
}

// ----- 响应 -----

// TokenResponse Token 响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	User      *User  `json:"user"`
}

// UserResponse 用户响应
type UserResponse struct {
	User *User `json:"user"`
}

// ConversationResponse 对话响应
type ConversationResponse struct {
	Conversation *Conversation `json:"conversation"`
}

// ConversationListResponse 对话列表响应
type ConversationListResponse struct {
	Conversations []*Conversation `json:"conversations"`
	Total         int64           `json:"total"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Content   string                 `json:"content"`
	Thinking  string                 `json:"thinking,omitempty"`
	Features  map[string]interface{} `json:"features,omitempty"`
	Model     string                `json:"model"`
	Tokens    int                   `json:"tokens,omitempty"`
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Features map[string]interface{} `json:"features"`
}

// ProvidersResponse LLM 提供商列表
type ProvidersResponse struct {
	Providers []string `json:"providers"`
}

// HealthResponse 健康检查
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// ErrorResponse 错误
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}
