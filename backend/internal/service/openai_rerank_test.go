package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExtractOpenAIRerankUsage(t *testing.T) {
	body := []byte(`{
		"results": [
			{"index": 0, "relevance_score": 0.92},
			{"index": 1, "relevance_score": 0.78}
		],
		"usage": {"prompt_tokens": 10, "total_tokens": 10}
	}`)

	usage := extractOpenAIRerankUsage(body)
	if usage.InputTokens != 10 || usage.OutputTokens != 0 {
		t.Fatalf("usage = %+v, want input=10 output=0", usage)
	}
}

func TestExtractOpenAIRerankUsageMissingUsage(t *testing.T) {
	body := []byte(`{"results":[{"index":0,"relevance_score":0.92}]}`)

	usage := extractOpenAIRerankUsage(body)
	if usage != (OpenAIUsage{}) {
		t.Fatalf("usage = %+v, want zero value", usage)
	}
}

func TestForwardRerankExternalPassthroughPreservesRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/rerank?trace=abc", nil)
	MarkExternalOpenAIRequestPassthrough(c, true)

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"req-rerank"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"results":[{"index":1,"relevance_score":0.78},{"index":0,"relevance_score":0.92}],"usage":{"prompt_tokens":10,"total_tokens":10}}`))),
	}}
	svc := &OpenAIGatewayService{
		httpUpstream: upstream,
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
	}
	account := &Account{
		ID:       123,
		Platform: PlatformExternalOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":                     "sk-test",
			"base_url":                    "https://example.test/api/v3",
			"request_passthrough_enabled": true,
			"model_mapping": map[string]any{
				"rerank-local": "rerank-upstream",
			},
			"endpoint_paths": map[string]any{
				"rerank": "/rerank",
			},
		},
	}
	body := []byte(`{"model":"rerank-local","query":"hello","documents":["a","b"],"unknown":{"keep":true}}`)

	result, err := svc.ForwardRerank(context.Background(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "https://example.test/api/v3/rerank?trace=abc", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-test", upstream.lastReq.Header.Get("Authorization"))
	require.JSONEq(t, string(body), string(upstream.lastBody))
	require.Equal(t, "rerank-local", result.UpstreamModel)
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"results":[{"index":1,"relevance_score":0.78},{"index":0,"relevance_score":0.92}],"usage":{"prompt_tokens":10,"total_tokens":10}}`, rec.Body.String())
}
