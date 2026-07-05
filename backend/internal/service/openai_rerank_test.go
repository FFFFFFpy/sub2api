package service

import "testing"

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
