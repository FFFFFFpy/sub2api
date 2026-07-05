package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type ExternalOpenAIEndpoint string

const (
	ExternalEndpointChatCompletions ExternalOpenAIEndpoint = "chat_completions"
	ExternalEndpointResponses       ExternalOpenAIEndpoint = "responses"
	ExternalEndpointEmbeddings      ExternalOpenAIEndpoint = "embeddings"
	ExternalEndpointRerank          ExternalOpenAIEndpoint = "rerank"
)

func trimTrailingSlash(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "/")
}

func joinURLPath(base, path string) string {
	normalizedBase := trimTrailingSlash(base)
	normalizedPath := "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	if normalizedPath == "/" {
		return normalizedBase
	}
	if normalizedBase == "" {
		return normalizedPath
	}
	if strings.HasSuffix(normalizedBase, normalizedPath) {
		return normalizedBase
	}
	return normalizedBase + normalizedPath
}

func defaultExternalOpenAIEndpointPath(endpoint ExternalOpenAIEndpoint) string {
	switch endpoint {
	case ExternalEndpointChatCompletions:
		return "/chat/completions"
	case ExternalEndpointResponses:
		return "/responses"
	case ExternalEndpointEmbeddings:
		return "/embeddings"
	case ExternalEndpointRerank:
		return "/rerank"
	default:
		return ""
	}
}

func externalOpenAIEndpointFromCapability(capability OpenAIEndpointCapability) ExternalOpenAIEndpoint {
	switch capability {
	case OpenAIEndpointCapabilityChatCompletions:
		return ExternalEndpointChatCompletions
	case OpenAIEndpointCapabilityResponses:
		return ExternalEndpointResponses
	case OpenAIEndpointCapabilityEmbeddings:
		return ExternalEndpointEmbeddings
	case OpenAIEndpointCapabilityRerank:
		return ExternalEndpointRerank
	default:
		return ""
	}
}

func credentialStringMap(credentials map[string]any, key string) map[string]string {
	if credentials == nil {
		return nil
	}
	raw, ok := credentials[key]
	if !ok || raw == nil {
		return nil
	}
	out := map[string]string{}
	switch m := raw.(type) {
	case map[string]string:
		for k, v := range m {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	case map[string]any:
		for k, v := range m {
			if s, ok := v.(string); ok {
				out[strings.TrimSpace(k)] = strings.TrimSpace(s)
			}
		}
	}
	return out
}

func splitPathAndQuery(path string) (string, string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", ""
	}
	if i := strings.Index(path, "?"); i >= 0 {
		return path[:i], path[i+1:]
	}
	return path, ""
}

func appendRawQuery(target, rawQuery string) string {
	rawQuery = strings.TrimSpace(rawQuery)
	if rawQuery == "" {
		return target
	}
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	return target + separator + rawQuery
}

func buildExternalOpenAICompatibleURL(baseURL string, endpointBaseURLs map[string]string, endpointPaths map[string]string, endpoint ExternalOpenAIEndpoint, incomingPath string) (string, error) {
	endpointKey := string(endpoint)
	selectedBase := strings.TrimSpace(endpointBaseURLs[endpointKey])
	if selectedBase == "" {
		selectedBase = strings.TrimSpace(baseURL)
	}
	if selectedBase == "" {
		return "", fmt.Errorf("base_url is required")
	}

	path := strings.TrimSpace(endpointPaths[endpointKey])
	if path == "" {
		path = defaultExternalOpenAIEndpointPath(endpoint)
	}
	if path == "" {
		return "", fmt.Errorf("unsupported external OpenAI endpoint: %s", endpoint)
	}

	path, configuredQuery := splitPathAndQuery(path)
	_, incomingQuery := splitPathAndQuery(incomingPath)
	target := joinURLPath(selectedBase, path)
	if configuredQuery != "" {
		target = appendRawQuery(target, configuredQuery)
	}
	if incomingQuery != "" {
		target = appendRawQuery(target, incomingQuery)
	}
	return target, nil
}

func (s *OpenAIGatewayService) externalOpenAICompatibleURL(account *Account, endpoint ExternalOpenAIEndpoint, incomingPath string) (string, error) {
	if account == nil || !account.IsExternalOpenAICompatibleAPIKey() {
		return "", fmt.Errorf("external OpenAI-compatible API key account is required")
	}
	targetURL, err := buildExternalOpenAICompatibleURL(
		account.GetCredential("base_url"),
		credentialStringMap(account.Credentials, "endpoint_base_urls"),
		credentialStringMap(account.Credentials, "endpoint_paths"),
		endpoint,
		incomingPath,
	)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	baseForValidation := parsed.Scheme + "://" + parsed.Host
	if parsed.Path != "" {
		baseForValidation = joinURLPath(baseForValidation, strings.TrimRight(parsed.Path, "/"))
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseForValidation)
	if err != nil {
		return "", err
	}
	if parsed.RawQuery != "" {
		return appendRawQuery(validatedURL, parsed.RawQuery), nil
	}
	return validatedURL, nil
}

func externalOpenAIIncomingPath(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.RequestURI()
}

func (s *OpenAIGatewayService) externalOpenAICompatibleResponsesURL(account *Account, incomingPath string) (string, bool, error) {
	if account == nil || !account.IsExternalOpenAICompatibleAPIKey() {
		return "", false, nil
	}
	targetURL, err := s.externalOpenAICompatibleURL(account, ExternalEndpointResponses, incomingPath)
	return targetURL, true, err
}

const externalOpenAIRequestPassthroughContextKey = "external_openai_request_passthrough"

func MarkExternalOpenAIRequestPassthrough(c *gin.Context, enabled bool) {
	if c != nil {
		c.Set(externalOpenAIRequestPassthroughContextKey, enabled)
	}
}

func externalOpenAIRequestPassthroughEnabled(c *gin.Context, account *Account) bool {
	if account == nil || !account.IsExternalOpenAICompatible() {
		return false
	}
	if c != nil {
		if raw, exists := c.Get(externalOpenAIRequestPassthroughContextKey); exists {
			if enabled, ok := raw.(bool); ok {
				return enabled
			}
		}
	}
	return account.IsExternalOpenAIRequestPassthroughEnabled(nil)
}
