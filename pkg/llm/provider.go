package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ChatRequest struct {
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Temp      float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
	Model    string `json:"model"`
	Tokens   int    `json:"tokens,omitempty"`
}

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Name() string
	SupportsThinking() bool
}

type Factory struct {
	providers map[string]Provider
	defaultP  string
}

func NewFactory() *Factory {
	return &Factory{providers: make(map[string]Provider), defaultP: "coze"}
}
func (f *Factory) Register(name string, p Provider) { f.providers[name] = p }
func (f *Factory) SetDefault(name string)           { f.defaultP = name }
func (f *Factory) List() []string {
	ns := make([]string, 0, len(f.providers))
	for n := range f.providers { ns = append(ns, n) }
	return ns
}
func (f *Factory) Get(name string) (Provider, error) {
	if name == "" { name = f.defaultP }
	if p, ok := f.providers[name]; ok { return p, nil }
	return nil, fmt.Errorf("provider not found: %s", name)
}

// ----- EdgeFn (白山智算) -----

type EdgeFnProvider struct{ APIKey, BaseURL, Model string }

type efResp struct {
	Choices []struct {
		Message struct {
			Content          *string `json:"content"`
			ReasoningContent *string `json:"reasoning_content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewEdgeFnProvider(apiKey, baseURL, model string) *EdgeFnProvider {
	if baseURL == "" { baseURL = "https://api.edgefn.net/v1" }
	return &EdgeFnProvider{APIKey: apiKey, BaseURL: strings.TrimSuffix(baseURL, "/"), Model: model}
}
func (*EdgeFnProvider) Name() string           { return "edgefn" }
func (*EdgeFnProvider) SupportsThinking() bool { return true }

func (p *EdgeFnProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := p.Model
	if req.Model != "" && req.Model != "edgefn" { model = req.Model }
	maxTokens := req.MaxTokens
	if maxTokens == 0 { maxTokens = 2000 }
	body, err := doHTTP(ctx, "POST", p.BaseURL+"/chat/completions",
		map[string]interface{}{"model": model, "messages": req.Messages, "max_tokens": maxTokens},
		map[string]string{"Authorization": "Bearer " + p.APIKey}, 120)
	if err != nil { return nil, fmt.Errorf("edgefn call failed: %w", err) }
	var resp efResp
	if err := json.Unmarshal(body, &resp); err != nil { return nil, fmt.Errorf("edgefn parse failed: %w", err) }
	if len(resp.Choices) == 0 { return nil, fmt.Errorf("edgefn: no choices") }
	msg := resp.Choices[0].Message
	content := ""
	thinking := ""
	if msg.Content != nil { content = *msg.Content }
	if msg.ReasoningContent != nil { thinking = *msg.ReasoningContent }
	if content == "" && thinking != "" { content = thinking }
	if content == "" { content = "[模型未返回有效内容]" }
	return &ChatResponse{Content: content, Thinking: thinking, Model: model}, nil
}

// ----- Coze -----

type CozeProvider struct{ APIKey, BotID, BaseURL string }

type cozeResp struct {
	Messages []struct {
		Role, Content, Type string
	}
}

func NewCozeProvider(apiKey, botID, baseURL string) *CozeProvider {
	if baseURL == "" { baseURL = "https://api.coze.cn" }
	return &CozeProvider{APIKey: apiKey, BotID: botID, BaseURL: baseURL}
}
func (*CozeProvider) Name() string           { return "coze" }
func (*CozeProvider) SupportsThinking() bool { return true }

func (p *CozeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages { msgs[i] = map[string]string{"role": m.Role, "content": m.Content} }
	body, err := doHTTP(ctx, "POST", p.BaseURL+"/v3/chat",
		map[string]interface{}{"bot_id": p.BotID, "messages": msgs, "stream": false},
		map[string]string{"Authorization": "Bearer " + p.APIKey}, 60)
	if err != nil { return nil, err }
	var resp cozeResp
	if err := json.Unmarshal(body, &resp); err != nil { return nil, err }
	content := ""
	for _, msg := range resp.Messages {
		if msg.Role == "assistant" && msg.Type == "answer" { content += msg.Content }
	}
	return &ChatResponse{Content: content, Model: "coze"}, nil
}

// ----- OpenAI -----

type OpenAIProvider struct{ APIKey, BaseURL string }

type oaResp struct {
	Choices []struct{ Message struct{ Content string } }
}

func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	if baseURL == "" { baseURL = "https://api.openai.com/v1" }
	return &OpenAIProvider{APIKey: apiKey, BaseURL: strings.TrimSuffix(baseURL, "/")}
}
func (*OpenAIProvider) Name() string           { return "openai" }
func (*OpenAIProvider) SupportsThinking() bool { return false }

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" { model = "gpt-4o" }
	temp := req.Temp
	if temp == 0 { temp = 0.7 }
	body, err := doHTTP(ctx, "POST", p.BaseURL+"/chat/completions",
		map[string]interface{}{"model": model, "messages": req.Messages, "max_tokens": req.MaxTokens, "temperature": temp},
		map[string]string{"Authorization": "Bearer " + p.APIKey}, 120)
	if err != nil { return nil, err }
	var resp oaResp
	if err := json.Unmarshal(body, &resp); err != nil { return nil, err }
	if len(resp.Choices) == 0 { return nil, fmt.Errorf("no choices") }
	return &ChatResponse{Content: resp.Choices[0].Message.Content, Model: model}, nil
}

// ----- Proxy -----

type ProxyProvider struct{ URL string }

type pxResp struct {
	Content  string
	Thinking string `json:"thinking,omitempty"`
}

func NewProxyProvider(url string) *ProxyProvider { return &ProxyProvider{URL: url} }
func (*ProxyProvider) Name() string              { return "proxy" }
func (*ProxyProvider) SupportsThinking() bool     { return true }

func (p *ProxyProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := doHTTP(ctx, "POST", p.URL+"/chat", map[string]interface{}{"messages": req.Messages}, nil, 120)
	if err != nil { return nil, err }
	var resp pxResp
	if err := json.Unmarshal(body, &resp); err != nil { return nil, err }
	return &ChatResponse{Content: resp.Content, Thinking: resp.Thinking, Model: "proxy"}, nil
}

// ----- HTTP -----

func doHTTP(ctx context.Context, method, url string, body interface{}, headers map[string]string, timeout int) ([]byte, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil { return nil, err }
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil { return nil, err }
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers { req.Header.Set(k, v) }
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(out)) }
	return out, nil
}
