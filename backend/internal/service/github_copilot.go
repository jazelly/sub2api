package service

import (
	"net/http"
	"strings"
)

// DefaultGitHubCopilotBaseURL is the Copilot inference API. Its endpoints have
// no /v1 prefix: Chat Completions is {base}/chat/completions.
const DefaultGitHubCopilotBaseURL = "https://api.githubcopilot.com"

// GitHub Copilot accounts store a GitHub OAuth token (gho_...) as api_key. The
// Copilot API accepts it directly as the bearer credential together with the
// editor identification headers below.
var githubCopilotDefaultHeaders = map[string]string{
	"Editor-Version":         "vscode/1.99.3",
	"Editor-Plugin-Version":  "copilot-chat/0.26.7",
	"Copilot-Integration-Id": "vscode-chat",
	"Openai-Intent":          "conversation-edits",
}

// IsGitHubCopilot reports whether the account relays a GitHub Copilot seat.
func (a *Account) IsGitHubCopilot() bool {
	return a != nil && a.Platform == PlatformGitHubCopilot
}

func applyGitHubCopilotHeaders(account *Account, header http.Header) {
	if !account.IsGitHubCopilot() {
		return
	}
	for key, value := range githubCopilotDefaultHeaders {
		header.Set(key, value)
	}
}

func buildGitHubCopilotChatCompletionsURL(base string) string {
	return strings.TrimRight(strings.TrimSpace(base), "/") + "/chat/completions"
}

// githubCopilotModelPrefixes are the provider prefixes a composite group
// accepts for Copilot models. Copilot serves model IDs that collide with other
// platforms (gpt-*, claude-*, gemini-*), so clients select Copilot explicitly,
// e.g. github-copilot/claude-sonnet-5.
var githubCopilotModelPrefixes = []string{"github-copilot", "github_copilot", "copilot"}

func isGitHubCopilotModelPrefix(provider string) bool {
	for _, prefix := range githubCopilotModelPrefixes {
		if provider == prefix {
			return true
		}
	}
	return false
}

// stripGitHubCopilotModelPrefix returns the Copilot upstream model ID for a
// prefixed public model, e.g. github-copilot/gpt-4.1 -> gpt-4.1.
func stripGitHubCopilotModelPrefix(model string) string {
	trimmed := strings.TrimSpace(model)
	slash := strings.IndexByte(trimmed, '/')
	if slash <= 0 || !isGitHubCopilotModelPrefix(strings.ToLower(trimmed[:slash])) {
		return trimmed
	}
	return strings.TrimSpace(trimmed[slash+1:])
}
