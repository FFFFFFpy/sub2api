package service

import "testing"

func TestBuildExternalOpenAICompatibleURL(t *testing.T) {
	tests := []struct {
		name             string
		baseURL          string
		endpointBaseURLs map[string]string
		endpointPaths    map[string]string
		endpoint         ExternalOpenAIEndpoint
		incomingPath     string
		want             string
	}{
		{
			name:     "ark chat",
			baseURL:  "https://ark.cn-beijing.volces.com/api/coding/v3",
			endpoint: ExternalEndpointChatCompletions,
			want:     "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions",
		},
		{
			name:     "ark responses",
			baseURL:  "https://ark.cn-beijing.volces.com/api/coding/v3",
			endpoint: ExternalEndpointResponses,
			want:     "https://ark.cn-beijing.volces.com/api/coding/v3/responses",
		},
		{
			name:     "ark embeddings",
			baseURL:  "https://ark.cn-beijing.volces.com/api/coding/v3",
			endpoint: ExternalEndpointEmbeddings,
			want:     "https://ark.cn-beijing.volces.com/api/coding/v3/embeddings",
		},
		{
			name:     "ark rerank",
			baseURL:  "https://ark.cn-beijing.volces.com/api/coding/v3",
			endpoint: ExternalEndpointRerank,
			want:     "https://ark.cn-beijing.volces.com/api/coding/v3/rerank",
		},
		{
			name:    "xunfei responses endpoint base override",
			baseURL: "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2",
			endpointBaseURLs: map[string]string{
				"responses": "https://maas-coding-api.cn-huabei-1.xf-yun.com/v1",
			},
			endpoint: ExternalEndpointResponses,
			want:     "https://maas-coding-api.cn-huabei-1.xf-yun.com/v1/responses",
		},
		{
			name:    "xunfei maas embeddings endpoint base override",
			baseURL: "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2",
			endpointBaseURLs: map[string]string{
				"embeddings": "https://maas-api.cn-huabei-1.xf-yun.com/v2",
			},
			endpoint: ExternalEndpointEmbeddings,
			want:     "https://maas-api.cn-huabei-1.xf-yun.com/v2/embeddings",
		},
		{
			name:     "custom endpoint path and incoming query",
			baseURL:  "https://example.com/openai/v3/",
			endpoint: ExternalEndpointRerank,
			endpointPaths: map[string]string{
				"rerank": "ranking/rerank?configured=1",
			},
			incomingPath: "/v1/rerank?trace=abc",
			want:         "https://example.com/openai/v3/ranking/rerank?configured=1&trace=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildExternalOpenAICompatibleURL(tt.baseURL, tt.endpointBaseURLs, tt.endpointPaths, tt.endpoint, tt.incomingPath)
			if err != nil {
				t.Fatalf("buildExternalOpenAICompatibleURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildExternalOpenAICompatibleURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
