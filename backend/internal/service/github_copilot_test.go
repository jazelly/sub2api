//go:build unit

package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func githubCopilotTestAccount() *Account {
	return &Account{
		ID:          301,
		Name:        "copilot",
		Platform:    PlatformGitHubCopilot,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "gho_test"},
	}
}

func requireGitHubCopilotUpstreamRequest(t *testing.T, req *http.Request) {
	t.Helper()
	require.Equal(t, "https://api.githubcopilot.com/chat/completions", req.URL.String())
	require.Equal(t, "Bearer gho_test", req.Header.Get("Authorization"))
	require.Equal(t, "vscode-chat", req.Header.Get("Copilot-Integration-Id"))
	require.Equal(t, "vscode/1.99.3", req.Header.Get("Editor-Version"))
	require.Equal(t, "copilot-chat/0.26.7", req.Header.Get("Editor-Plugin-Version"))
}

func TestGitHubCopilotForwardsChatCompletions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, stream := range []bool{false, true} {
		upstream := &chatRoleContractUpstream{}
		svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
		body := []byte(`{"model":"gpt-4.1","messages":[{"role":"user","content":"ping"}],"stream":false}`)
		if stream {
			body = []byte(`{"model":"gpt-4.1","messages":[{"role":"user","content":"ping"}],"stream":true}`)
		}

		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
		result, err := svc.ForwardAsChatCompletions(context.Background(), c, githubCopilotTestAccount(), body, "", "")

		require.NoError(t, err)
		requireGitHubCopilotUpstreamRequest(t, upstream.lastReq)
		require.Equal(t, "gpt-4.1", gjson.GetBytes(upstream.lastBody, "model").String())
		require.Equal(t, 9, result.Usage.InputTokens)
		require.Equal(t, 1, result.Usage.OutputTokens)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), "pong")
	}
}

func TestGitHubCopilotForwardsResponsesViaChatCompletions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &chatRoleContractUpstream{}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	body := []byte(`{"model":"gpt-4.1","input":"ping","stream":false}`)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	result, err := svc.Forward(context.Background(), c, githubCopilotTestAccount(), body)

	require.NoError(t, err)
	requireGitHubCopilotUpstreamRequest(t, upstream.lastReq)
	require.Equal(t, "ping", gjson.GetBytes(upstream.lastBody, "messages.0.content").String())
	require.Equal(t, 9, result.Usage.InputTokens)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "pong")
}

func TestCompositeRoutesPrefixedModelsToGitHubCopilot(t *testing.T) {
	for _, model := range []string{"github-copilot/claude-sonnet-5", "copilot/claude-sonnet-5"} {
		decision, err := NewCompositeRouteResolver(nil).Resolve(context.Background(), 1, model, "")
		require.NoError(t, err)
		require.True(t, decision.Matched)
		require.Equal(t, PlatformGitHubCopilot, decision.TargetPlatform)
		require.Equal(t, "claude-sonnet-5", decision.UpstreamModel)
	}

	decision, err := NewCompositeRouteResolver(nil).Resolve(context.Background(), 1, "claude-sonnet-5", "")
	require.NoError(t, err)
	require.Equal(t, PlatformAnthropic, decision.TargetPlatform)
}

func TestPricingResolvesCopilotDottedClaudeVersions(t *testing.T) {
	sonnet := &LiteLLMModelPricing{InputCostPerToken: 2e-6}
	haiku3 := &LiteLLMModelPricing{InputCostPerToken: 2.5e-7}
	svc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"claude-sonnet-5-5": sonnet,
		"claude-3-haiku":    haiku3,
	}}

	require.Same(t, sonnet, svc.GetModelPricing("claude-sonnet-5.5"))
	require.Same(t, sonnet, svc.GetModelPricing("github-copilot/claude-sonnet-5.5"))
}
