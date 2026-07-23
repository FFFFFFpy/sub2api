#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path

ROOT = Path.cwd()


def read(rel: str) -> str:
    return (ROOT / rel).read_text(encoding="utf-8")


def write(rel: str, text: str) -> None:
    (ROOT / rel).write_text(text, encoding="utf-8")


def replace_exact(rel: str, old: str, new: str, expected: int = 1) -> None:
    text = read(rel)
    count = text.count(old)
    if count != expected:
        raise RuntimeError(f"{rel}: expected {expected} matches, found {count}: {old[:120]!r}")
    write(rel, text.replace(old, new))


# Keep native OpenAI media routes isolated while allowing external-compatible
# providers on the embeddings and rerank surfaces.
replace_exact(
    "backend/internal/server/routes/gateway.go",
    '''\tisEmbeddingsGatewayPlatform := func(c *gin.Context) bool {\n''',
    '''\tisOpenAIOnlyEndpointGatewayPlatform := func(c *gin.Context) bool {\n\t\treturn getGroupPlatform(c) == service.PlatformOpenAI\n\t}\n\tisEmbeddingsGatewayPlatform := func(c *gin.Context) bool {\n''',
)
replace_exact(
    "backend/internal/server/routes/gateway.go",
    '''\t\tif !isOpenAIOnlyEndpointGatewayPlatform(c) {\n\t\t\tservice.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)\n\t\t\tc.JSON(http.StatusNotFound, gin.H{\n\t\t\t\t"error": gin.H{\n\t\t\t\t\t"type":    "not_found_error",\n\t\t\t\t\t"message": "Embeddings API is not supported for this platform",\n\t\t\t\t},\n\t\t\t})\n\t\t\treturn\n\t\t}\n\t\th.OpenAIGateway.Embeddings(c)\n\t})\n\tr.POST("/images/generations",''',
    '''\t\tif !isEmbeddingsGatewayPlatform(c) {\n\t\t\tservice.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)\n\t\t\tc.JSON(http.StatusNotFound, gin.H{\n\t\t\t\t"error": gin.H{\n\t\t\t\t\t"type":    "not_found_error",\n\t\t\t\t\t"message": "Embeddings API is not supported for this platform",\n\t\t\t\t},\n\t\t\t})\n\t\t\treturn\n\t\t}\n\t\th.OpenAIGateway.Embeddings(c)\n\t})\n\tr.POST("/rerank", textBodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gin.HandlerFunc(apiKeyAuth), compositeTarget, requireGroupAnthropic, rerankHandler)\n\tr.POST("/images/generations",''',
)

# Rerank carries user prompt-like documents/query data and must pass the same
# security audit coordinator as other model-executing endpoints.
replace_exact(
    "backend/internal/handler/openai_rerank.go",
    '''\tsetOpsRequestContext(c, reqModel, false)\n\tsetOpsEndpointContext(c, "", int16(service.RequestTypeSync))\n\n\tchannelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)\n''',
    '''\tsetOpsRequestContext(c, reqModel, false)\n\tsetOpsEndpointContext(c, "", int16(service.RequestTypeSync))\n\tif decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, "openai_rerank", reqModel, body); decision != nil && !decision.AllowNextStage {\n\t\th.openAISecurityAuditError(c, decision)\n\t\treturn\n\t}\n\n\tchannelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)\n''',
)
replace_exact(
    "backend/internal/server/routes/prompt_audit_route_coverage_test.go",
    '''\t\t"/embeddings":               {"openai_embeddings.go"},\n\t\t"/alpha/search":             {"openai_alpha_search.go"},\n''',
    '''\t\t"/embeddings":               {"openai_embeddings.go"},\n\t\t"/rerank":                  {"openai_rerank.go"},\n\t\t"/alpha/search":             {"openai_alpha_search.go"},\n''',
)

# Preserve request-time provider classification in both no-account paths.
replace_exact(
    "backend/internal/handler/openai_chat_completions.go",
    '''classifyOpenAICompatibleNoAccountErrorFromGin(c, h.gatewayService, apiKey, reqModel, reqModel)''',
    '''classifyNoAccountErrorFromGin(c, h.gatewayService, apiKey, reqModel, reqModel, requestPlatform)''',
    expected=2,
)

# True passthrough means the request body is not enriched with stream_options;
# external providers may legitimately omit usage in their final SSE chunk.
replace_exact(
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    '''\tif clientStream {\n\t\tvar usageErr error\n\t\tupstreamBody, usageErr = ensureOpenAIChatStreamUsage(upstreamBody)\n''',
    '''\tif clientStream && !requestPassthrough {\n\t\tvar usageErr error\n\t\tupstreamBody, usageErr = ensureOpenAIChatStreamUsage(upstreamBody)\n''',
)

# Update API contract snapshots for the intentionally exposed group flag and
# the new quota platform. These are response-contract changes, not test bypasses.
replace_exact(
    "backend/internal/server/api_contract_test.go",
    '''\t\t\t\t\t\t"reasoning_effort_mappings": null,\n\t\t\t\t\t\t"rpm_limit": 0,\n''',
    '''\t\t\t\t\t\t"reasoning_effort_mappings": null,\n\t\t\t\t\t\t"request_passthrough_enabled": false,\n\t\t\t\t\t\t"rpm_limit": 0,\n''',
)
old_quota = '''"default_platform_quotas": {"anthropic":{"daily":null,"weekly":null,"monthly":null},"antigravity":{"daily":null,"weekly":null,"monthly":null},"gemini":{"daily":null,"weekly":null,"monthly":null},"grok":{"daily":null,"weekly":null,"monthly":null},"openai":{"daily":null,"weekly":null,"monthly":null}}'''
new_quota = '''"default_platform_quotas": {"anthropic":{"daily":null,"weekly":null,"monthly":null},"antigravity":{"daily":null,"weekly":null,"monthly":null},"external_openai_compatible":{"daily":null,"weekly":null,"monthly":null},"gemini":{"daily":null,"weekly":null,"monthly":null},"grok":{"daily":null,"weekly":null,"monthly":null},"openai":{"daily":null,"weekly":null,"monthly":null}}'''
replace_exact("backend/internal/server/api_contract_test.go", old_quota, new_quota, expected=2)

replace_exact(
    "frontend/src/types/index.ts",
    "export type GroupPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok' | 'composite'",
    "export type GroupPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok' | 'composite' | 'external_openai_compatible' | 'volcengine_coding' | 'xunfei_coding'",
)

modified_files = [
    "backend/internal/server/routes/gateway.go",
    "backend/internal/handler/openai_rerank.go",
    "backend/internal/server/routes/prompt_audit_route_coverage_test.go",
    "backend/internal/handler/openai_chat_completions.go",
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    "backend/internal/server/api_contract_test.go",
    "frontend/src/types/index.ts",
]
subprocess.run(["gofmt", "-w", *[path for path in modified_files if path.endswith(".go")]], check=True)
subprocess.run(["git", "add", *modified_files], check=True)
