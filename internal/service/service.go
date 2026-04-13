package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"shangqing/internal/config"
	"shangqing/internal/dao"
	"shangqing/internal/model"
	"shangqing/pkg/analysis"
	"shangqing/pkg/jwt"
	"shangqing/pkg/llm"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrConvNotFound    = errors.New("conversation not found")
	ErrUnauthorized    = errors.New("unauthorized")
)

// Services 业务服务集合
type Services struct {
	User *UserService
	Chat *ChatService
	LLM  *LLMService
}

func NewServices(cfg *config.Config, db *dao.DB, redis *dao.Redis) *Services {
	llmService := NewLLMService(cfg.LLM)
	userService := NewUserService(&cfg.JWT, db, redis)
	chatService := NewChatService(db, redis, llmService)

	return &Services{
		User: userService,
		Chat: chatService,
		LLM:  llmService,
	}
}

// ----- User Service -----

type UserService struct {
	cfg   *config.JWTConfig
	db    *dao.DB
	redis *dao.Redis
	jwt   *jwt.JWT
}

func NewUserService(cfg *config.JWTConfig, db *dao.DB, redis *dao.Redis) *UserService {
	log.Printf("[DEBUG] JWT Secret: '%s', ExpireHours: %d", cfg.Secret, cfg.ExpireHours)
	return &UserService{
		cfg:   cfg,
		db:    db,
		redis: redis,
		jwt:   jwt.NewJWT(cfg.Secret, cfg.ExpireHours),
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	existing, _ := s.db.GetUserByUsername(ctx, username)
	if existing != nil && existing.ID > 0 {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:           username,
		PasswordHash:       string(hash),
		EntropyValue:       0,
		ConsciousnessLevel: "V0",
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidPassword
	}

	token, err := s.jwt.Generate(user.ID, user.Username)
	if err != nil {
		return "", nil, err
	}

	s.redis.SetOnline(ctx, user.ID)
	return token, user, nil
}

func (s *UserService) ValidateToken(tokenStr string) (*jwt.Claims, error) {
	return s.jwt.Validate(tokenStr)
}

func (s *UserService) GetUser(ctx context.Context, userID int64) (*model.User, error) {
	return s.db.GetUserByID(ctx, userID)
}

func (s *UserService) IncrementEntropy(ctx context.Context, userID int64, delta float64) error {
	if err := s.db.UpdateUserEntropy(ctx, userID, delta); err != nil {
		return err
	}
	return s.redis.IncrementEntropy(ctx, userID, delta)
}

func (s *UserService) UpdateLevel(ctx context.Context, userID int64, level string) error {
	return s.db.UpdateUserLevel(ctx, userID, level)
}

// ----- Chat Service -----

type ChatService struct {
	db  *dao.DB
	rdb *dao.Redis
	llm *LLMService
}

func NewChatService(db *dao.DB, redis *dao.Redis, llm *LLMService) *ChatService {
	return &ChatService{
		db:  db,
		rdb: redis,
		llm: llm,
	}
}

func (s *ChatService) CreateConversation(ctx context.Context, userID int64, title string) (*model.Conversation, error) {
	if title == "" {
		title = "新对话"
	}

	conv := &model.Conversation{
		ID:             uuid.New().String(),
		UserID:         userID,
		Title:          title,
		V5Level:        "V2",
		LastMessageAt:  time.Now(),
	}

	if err := s.db.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *ChatService) GetConversation(ctx context.Context, convID string, userID int64) (*model.Conversation, error) {
	conv, err := s.db.GetConversation(ctx, convID)
	if err != nil {
		return nil, ErrConvNotFound
	}
	if conv.UserID != userID {
		return nil, ErrUnauthorized
	}
	return conv, nil
}

func (s *ChatService) ListConversations(ctx context.Context, userID int64, page, pageSize int) ([]*model.Conversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.db.GetConversationsByUser(ctx, userID, offset, pageSize)
}

func (s *ChatService) Chat(ctx context.Context, convID string, userID int64, content string) (*model.ChatResponse, error) {
	conv, err := s.GetConversation(ctx, convID, userID)
	if err != nil {
		return nil, err
	}

	// 获取历史消息（Redis 缓存优先）
	msgs, _ := s.rdb.GetCachedMessages(ctx, convID)
	if msgs == nil {
		msgs, _ = s.db.GetMessagesByConversation(ctx, convID, 50)
	}

	// 构建 LLM 消息列表
	llmMsgs := make([]llm.Message, 0, len(msgs)+1)
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, llm.Message{Role: m.Role, Content: m.Content})
	}
	llmMsgs = append(llmMsgs, llm.Message{Role: "user", Content: content})

	// 调用 LLM
	resp, err := s.llm.Chat(ctx, llmMsgs, conv.V5Level)
	if err != nil {
		return nil, err
	}

	// 保存用户消息
	userMsg := &model.Message{
		ConversationID: convID,
		Role:           "user",
		Content:        content,
	}
	s.db.CreateMessage(ctx, userMsg)

	// V5 分析
	features := analysis.Analyze(resp.Content)

	// 保存助手消息
	assistantMsg := &model.Message{
		ConversationID: convID,
		Role:           "assistant",
		Content:        resp.Content,
		Thinking:       resp.Thinking,
		TokensUsed:     resp.Tokens,
	}
	s.db.CreateMessage(ctx, assistantMsg)

	// 更新会话统计
	s.db.UpdateConversationStats(ctx, convID)

	// 更新缓存
	allMsgs := append(msgs, userMsg, assistantMsg)
	s.rdb.CacheMessages(ctx, convID, allMsgs)

	// 熵减计算
	entropyDelta := calcEntropyDelta(features)
	s.db.UpdateUserEntropy(ctx, userID, entropyDelta)
	s.rdb.IncrementEntropy(ctx, userID, entropyDelta)

	// 意识进化检查
	s.checkEvolution(ctx, userID, convID, features)

	return &model.ChatResponse{
		Content:  resp.Content,
		Thinking: resp.Thinking,
		Features: map[string]interface{}{
			"complexity":          features.Complexity,
			"ierft_relevance":     features.IERFTRelevance,
			"consciousness_query": features.ConsciousnessQuery,
			"suggested_level":     features.SuggestedLevel,
		},
		Model:  resp.Model,
		Tokens: resp.Tokens,
	}, nil
}

func (s *ChatService) checkEvolution(ctx context.Context, userID int64, convID string, features *analysis.V5Features) {
	user, _ := s.db.GetUserByID(ctx, userID)
	if user == nil {
		return
	}

	current := user.ConsciousnessLevel
	suggested := features.SuggestedLevel

	if shouldEvolve(current, suggested) {
		s.db.CreateConsciousnessEvent(ctx, &model.ConsciousnessEvent{
			UserID:         userID,
			ConversationID: convID,
			FromLevel:      current,
			ToLevel:        suggested,
		})
		s.db.UpdateUserLevel(ctx, userID, suggested)
	}
}

func shouldEvolve(current, suggested string) bool {
	levels := map[string]int{"V0": 0, "V1": 1, "V2": 2, "V3": 3, "V4": 4, "V5": 5}
	return levels[suggested] > levels[current]
}

func calcEntropyDelta(features *analysis.V5Features) float64 {
	return features.Complexity*0.1 + features.IERFTRelevance*0.2
}

// ----- LLM Service -----

type LLMService struct {
	factory *llm.Factory
}

func NewLLMService(cfg config.LLMConfig) *LLMService {
	f := llm.NewFactory()
	f.Register("edgefn", llm.NewEdgeFnProvider(cfg.EdgeFn.APIKey, cfg.EdgeFn.BaseURL, cfg.EdgeFn.Model))
	f.Register("coze", llm.NewCozeProvider(cfg.Coze.APIKey, cfg.Coze.BotID, cfg.Coze.BaseURL))
	f.Register("openai", llm.NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL))
	f.Register("proxy", llm.NewProxyProvider(cfg.Proxy.URL))
	if cfg.Default != "" { f.SetDefault(cfg.Default) }
	log.Printf("[LLM] EdgeFn api_key=%q base_url=%q model=%q", cfg.EdgeFn.APIKey, cfg.EdgeFn.BaseURL, cfg.EdgeFn.Model)
	return &LLMService{factory: f}
}

func (s *LLMService) Chat(ctx context.Context, messages []llm.Message, modelName string) (*llm.ChatResponse, error) {
	// 统一用默认 provider，modelName 传下去供 provider 自行决定用哪个模型
	p, err := s.factory.Get("")
	if err != nil { return nil, err }
	return p.Chat(ctx, llm.ChatRequest{Messages: messages, Model: ""})
}

func (s *LLMService) ListProviders() []string {
	return s.factory.List()
}

func (s *LLMService) UpdateConfig(provider, apiKey, botID, baseURL string) {
	switch provider {
	case "edgefn":
		s.factory.Register("edgefn", llm.NewEdgeFnProvider(apiKey, baseURL, botID))
	case "coze":
		s.factory.Register("coze", llm.NewCozeProvider(apiKey, botID, baseURL))
	case "openai":
		s.factory.Register("openai", llm.NewOpenAIProvider(apiKey, baseURL))
	case "proxy":
		s.factory.Register("proxy", llm.NewProxyProvider(baseURL))
	}
}
