package service

import (
	"context"
	"testing"
)

func TestExternalOpenAICompatibleQuotaPlatformIsAllowed(t *testing.T) {
	if !IsAllowedQuotaPlatform(PlatformExternalOpenAI) {
		t.Fatalf("%s must be an allowed user platform quota platform", PlatformExternalOpenAI)
	}
}

func TestQuotaPlatformExternalOpenAICompatibleUsesGroupPlatform(t *testing.T) {
	apiKey := &APIKey{Group: &Group{Platform: PlatformExternalOpenAI}}
	if got := QuotaPlatform(context.Background(), apiKey); got != PlatformExternalOpenAI {
		t.Fatalf("QuotaPlatform external OpenAI-compatible = %q, want %q", got, PlatformExternalOpenAI)
	}
}
