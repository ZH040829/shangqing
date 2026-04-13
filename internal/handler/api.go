package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"shangqing/internal/middleware"
	"shangqing/internal/model"
	"shangqing/internal/service"
	"shangqing/pkg/analysis"
)

type Handler struct {
	svc *service.Services
}

func NewHandler(svc *service.Services) *Handler {
	return &Handler{svc: svc}
}

// ----- 健康检查 -----

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:  "ok",
		Service: "熵清 V5 Backend",
		Version: "1.0.0",
	})
}

// ----- 用户 -----

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.svc.User.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserExists {
			c.JSON(http.StatusConflict, model.ErrorResponse{Error: "user already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.UserResponse{User: user})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	token, user, err := h.svc.User.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, model.TokenResponse{
		Token:     token,
		ExpiresIn: 720 * 3600,
		User:      user,
	})
}

// GetProfile 获取个人资料
func (h *Handler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.svc.User.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, model.UserResponse{User: user})
}

// ----- 对话 -----

// CreateConversation 创建对话
func (h *Handler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req model.CreateConversationRequest
	c.ShouldBindJSON(&req) // 可选 body

	conv, err := h.svc.Chat.CreateConversation(c.Request.Context(), userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.ConversationResponse{Conversation: conv})
}

// ListConversations 对话列表
func (h *Handler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	convs, total, err := h.svc.Chat.ListConversations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.ConversationListResponse{
		Conversations: convs,
		Total:         total,
	})
}

// GetConversation 获取单个对话
func (h *Handler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	convID := c.Param("id")

	conv, err := h.svc.Chat.GetConversation(c.Request.Context(), convID, userID)
	if err != nil {
		if err == service.ErrConvNotFound || err == service.ErrUnauthorized {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "conversation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.ConversationResponse{Conversation: conv})
}

// Chat 对话
func (h *Handler) Chat(c *gin.Context) {
	userID := middleware.GetUserID(c)
	convID := c.Param("id")

	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.svc.Chat.Chat(c.Request.Context(), convID, userID, req.Content)
	if err != nil {
		if err == service.ErrConvNotFound || err == service.ErrUnauthorized {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "conversation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ----- 分析 -----

// Analyze 文本分析
func (h *Handler) Analyze(c *gin.Context) {
	var req model.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	// 复用 ChatService 的分析逻辑，但不需要对话上下文
	features := analyzeText(req.Text)

	c.JSON(http.StatusOK, model.AnalyzeResponse{Features: features})
}

// ----- LLM 配置 -----

// ListProviders 列出提供商
func (h *Handler) ListProviders(c *gin.Context) {
	providers := h.svc.LLM.ListProviders()
	c.JSON(http.StatusOK, model.ProvidersResponse{Providers: providers})
}

// UpdateLLMConfig 更新 LLM 配置
func (h *Handler) UpdateLLMConfig(c *gin.Context) {
	var req model.UpdateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	h.svc.LLM.UpdateConfig(req.Provider, req.APIKey, req.BotID, req.BaseURL)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DebugSecret 调试用（临时）
func (h *Handler) DebugSecret(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "check server log"})
}

// ----- 辅助函数 -----

func analyzeText(text string) map[string]interface{} {
	f := analysis.Analyze(text)
	return map[string]interface{}{
		"complexity":          f.Complexity,
		"ierft_relevance":     f.IERFTRelevance,
		"consciousness_query": f.ConsciousnessQuery,
		"suggested_level":    f.SuggestedLevel,
	}
}
