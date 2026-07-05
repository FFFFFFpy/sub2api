package service

import "testing"

func TestBuildVolcengineCodingURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		endpoint string
		want     string
	}{
		{"default chat", "", "/chat/completions", "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions"},
		{"default responses", "", "/responses", "https://ark.cn-beijing.volces.com/api/coding/v3/responses"},
		{"default embeddings", "", "/embeddings", "https://ark.cn-beijing.volces.com/api/coding/v3/embeddings"},
		{"default rerank", "", "/rerank", "https://ark.cn-beijing.volces.com/api/coding/v3/rerank"},
		{"root chat", "https://ark.cn-beijing.volces.com", "/chat/completions", "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions"},
		{"versioned chat", "https://ark.cn-beijing.volces.com/api/coding/v3", "/chat/completions", "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildVolcengineCodingURL(tt.base, tt.endpoint); got != tt.want {
				t.Fatalf("buildVolcengineCodingURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildXunfeiCodingURL(t *testing.T) {
	tests := []struct {
		name          string
		base          string
		embeddingBase string
		rerankBase    string
		kind          string
		want          string
	}{
		{"default chat", "", "", "", "chat", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/chat/completions"},
		{"default responses", "", "", "", "responses", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v1/responses"},
		{"default embeddings", "", "", "", "embeddings", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/embeddings"},
		{"default rerank", "", "", "", "rerank", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/rerank"},
		{"root chat", "https://maas-coding-api.cn-huabei-1.xf-yun.com", "", "", "chat", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/chat/completions"},
		{"versioned chat", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2", "", "", "chat", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/chat/completions"},
		{"versioned responses fixed v1", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2", "", "", "responses", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v1/responses"},
		{"versioned embeddings", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2", "", "", "embeddings", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/embeddings"},
		{"versioned rerank", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2", "", "", "rerank", "https://maas-coding-api.cn-huabei-1.xf-yun.com/v2/rerank"},
		{"maas embedding base", "", "https://maas-api.cn-huabei-1.xf-yun.com/v2", "", "embeddings", "https://maas-api.cn-huabei-1.xf-yun.com/v2/embeddings"},
		{"maas rerank base", "", "", "https://maas-api.cn-huabei-1.xf-yun.com/v2", "rerank", "https://maas-api.cn-huabei-1.xf-yun.com/v2/rerank"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.kind {
			case "chat":
				got = buildXunfeiCodingChatCompletionsURL(tt.base)
			case "responses":
				got = buildXunfeiCodingResponsesURL()
			case "embeddings":
				got = buildXunfeiCodingEmbeddingsURL(tt.base, tt.embeddingBase)
			case "rerank":
				got = buildXunfeiCodingRerankURL(tt.base, tt.embeddingBase, tt.rerankBase)
			}
			if got != tt.want {
				t.Fatalf("xunfei URL = %q, want %q", got, tt.want)
			}
		})
	}
}
