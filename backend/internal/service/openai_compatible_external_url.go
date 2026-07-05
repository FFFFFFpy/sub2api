package service

import (
	"net/url"
	"strings"
)

const (
	DefaultVolcengineCodingBaseURL = "https://ark.cn-beijing.volces.com/api/coding/v3"
	DefaultXunfeiCodingRootURL     = "https://maas-coding-api.cn-huabei-1.xf-yun.com"
	DefaultXunfeiCodingBaseURL     = DefaultXunfeiCodingRootURL + "/v2"
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

func ensurePathSuffix(base, suffix string) string {
	normalizedBase := trimTrailingSlash(base)
	normalizedSuffix := "/" + strings.Trim(strings.TrimSpace(suffix), "/")
	if normalizedBase == "" || normalizedSuffix == "/" {
		return normalizedBase
	}
	if parsed, err := url.Parse(normalizedBase); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		if strings.HasSuffix(strings.TrimRight(parsed.EscapedPath(), "/"), normalizedSuffix) ||
			strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), normalizedSuffix) {
			return normalizedBase
		}
	}
	if strings.HasSuffix(normalizedBase, normalizedSuffix) {
		return normalizedBase
	}
	return joinURLPath(normalizedBase, normalizedSuffix)
}

func volcengineCodingBaseURL(base string) string {
	if strings.TrimSpace(base) == "" {
		return DefaultVolcengineCodingBaseURL
	}
	return ensurePathSuffix(base, "/api/coding/v3")
}

func buildVolcengineCodingURL(base, endpoint string) string {
	return joinURLPath(volcengineCodingBaseURL(base), endpoint)
}

func xunfeiCodingBaseURL(base string) string {
	if strings.TrimSpace(base) == "" {
		return DefaultXunfeiCodingBaseURL
	}
	return ensurePathSuffix(base, "/v2")
}

func buildXunfeiCodingChatCompletionsURL(base string) string {
	return joinURLPath(xunfeiCodingBaseURL(base), "/chat/completions")
}

func buildXunfeiCodingResponsesURL() string {
	return joinURLPath(DefaultXunfeiCodingRootURL, "/v1/responses")
}

func buildXunfeiCodingEmbeddingsURL(base, embeddingBase string) string {
	selectedBase := strings.TrimSpace(embeddingBase)
	if selectedBase == "" {
		selectedBase = base
	}
	return joinURLPath(xunfeiCodingBaseURL(selectedBase), "/embeddings")
}

func buildXunfeiCodingRerankURL(base, embeddingBase, rerankBase string) string {
	selectedBase := strings.TrimSpace(rerankBase)
	if selectedBase == "" {
		selectedBase = strings.TrimSpace(embeddingBase)
	}
	if selectedBase == "" {
		selectedBase = base
	}
	return joinURLPath(xunfeiCodingBaseURL(selectedBase), "/rerank")
}

func (s *OpenAIGatewayService) externalOpenAICompatibleResponsesURL(account *Account) (string, bool, error) {
	if account == nil {
		return "", false, nil
	}
	switch account.Platform {
	case PlatformVolcengineCoding:
		baseURL := volcengineCodingBaseURL(account.GetCredential("base_url"))
		validatedURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return "", true, err
		}
		return buildVolcengineCodingURL(validatedURL, "/responses"), true, nil
	case PlatformXunfeiCoding:
		validatedURL, err := s.validateUpstreamBaseURL(DefaultXunfeiCodingRootURL)
		if err != nil {
			return "", true, err
		}
		return joinURLPath(validatedURL, "/v1/responses"), true, nil
	default:
		return "", false, nil
	}
}
