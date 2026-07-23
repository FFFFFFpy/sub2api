#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import subprocess
from pathlib import Path

ROOT = Path.cwd()


def read(rel: str) -> str:
    return (ROOT / rel).read_text(encoding="utf-8")


def write(rel: str, text: str) -> None:
    (ROOT / rel).write_text(text, encoding="utf-8")


def replace_once(rel: str, old: str, new: str) -> None:
    replace_exact(rel, old, new, 1)


def replace_exact(rel: str, old: str, new: str, expected: int) -> None:
    text = read(rel)
    count = text.count(old)
    if count != expected:
        raise RuntimeError(f"{rel}: expected {expected} matches, found {count}: {old[:100]!r}")
    write(rel, text.replace(old, new))


def regex_replace_once(rel: str, pattern: str, replacement: str, flags: int = 0) -> None:
    text = read(rel)
    updated, count = re.subn(pattern, replacement, text, count=1, flags=flags)
    if count != 1:
        raise RuntimeError(f"{rel}: expected exactly one regex match, found {count}: {pattern!r}")
    write(rel, updated)


# Restore the fork-specific provider identities alongside the upstream composite platform.
replace_once(
    "backend/internal/domain/constants.go",
    '''\tPlatformAnthropic   = "anthropic"\n\tPlatformOpenAI      = "openai"\n\tPlatformGemini      = "gemini"\n\tPlatformAntigravity = "antigravity"\n\tPlatformGrok        = "grok"\n\tPlatformComposite   = "composite"\n''',
    '''\tPlatformAnthropic        = "anthropic"\n\tPlatformOpenAI           = "openai"\n\tPlatformGemini           = "gemini"\n\tPlatformAntigravity      = "antigravity"\n\tPlatformGrok             = "grok"\n\tPlatformComposite        = "composite"\n\tPlatformExternalOpenAI   = "external_openai_compatible"\n\tPlatformVolcengineCoding = "volcengine_coding"\n\tPlatformXunfeiCoding     = "xunfei_coding"\n''',
)

replace_once(
    "backend/internal/service/domain_constants.go",
    '''\tPlatformAnthropic   = domain.PlatformAnthropic\n\tPlatformOpenAI      = domain.PlatformOpenAI\n\tPlatformGemini      = domain.PlatformGemini\n\tPlatformAntigravity = domain.PlatformAntigravity\n\tPlatformGrok        = domain.PlatformGrok\n\tPlatformComposite   = domain.PlatformComposite\n)\n''',
    '''\tPlatformAnthropic        = domain.PlatformAnthropic\n\tPlatformOpenAI           = domain.PlatformOpenAI\n\tPlatformGemini           = domain.PlatformGemini\n\tPlatformAntigravity      = domain.PlatformAntigravity\n\tPlatformGrok             = domain.PlatformGrok\n\tPlatformComposite        = domain.PlatformComposite\n\tPlatformExternalOpenAI   = domain.PlatformExternalOpenAI\n\tPlatformVolcengineCoding = domain.PlatformVolcengineCoding\n\tPlatformXunfeiCoding     = domain.PlatformXunfeiCoding\n)\n\n// NormalizeOpenAICompatiblePlatformForRouting folds the legacy dedicated coding\n// provider identifiers into the shared external OpenAI-compatible routing lane.\nfunc NormalizeOpenAICompatiblePlatformForRouting(platform string) string {\n\tswitch platform {\n\tcase PlatformVolcengineCoding, PlatformXunfeiCoding:\n\t\treturn PlatformExternalOpenAI\n\tdefault:\n\t\treturn platform\n\t}\n}\n''',
)

# Reconcile capability constants with upstream's newer capability set.
replace_once(
    "backend/internal/service/account.go",
    '''\tOpenAIEndpointCapabilityResponses       OpenAIEndpointCapability = "responses"\n\tOpenAIEndpointCapabilityEmbeddings      OpenAIEndpointCapability = "embeddings"\n\tOpenAIEndpointCapabilityAlphaSearch     OpenAIEndpointCapability = "alpha_search"\n''',
    '''\tOpenAIEndpointCapabilityResponses       OpenAIEndpointCapability = "responses"\n\tOpenAIEndpointCapabilityEmbeddings      OpenAIEndpointCapability = "embeddings"\n\tOpenAIEndpointCapabilityRerank          OpenAIEndpointCapability = "rerank"\n\tOpenAIEndpointCapabilityAlphaSearch     OpenAIEndpointCapability = "alpha_search"\n''',
)
regex_replace_once(
    "backend/internal/service/account.go",
    r'''\n\t// OpenAIEndpointCapabilityResponses 表示上游确实提供 /v1/responses 端点。\n(?:\t//.*\n){4}\tOpenAIEndpointCapabilityResponses OpenAIEndpointCapability = "responses"\n''',
    "\n",
)
replace_once(
    "backend/internal/service/account.go",
    '''func (a *Account) IsLegacyExternalOpenAICompatible() bool {\n\treturn a != nil && (a.IsVolcengineCoding() || a.IsXunfeiCoding())\n}\n\n''',
    '''func (a *Account) IsLegacyExternalOpenAICompatible() bool {\n\treturn a != nil && (a.IsVolcengineCoding() || a.IsXunfeiCoding())\n}\n\nfunc (a *Account) IsExternalOpenAICompatibleAPIKey() bool {\n\treturn a != nil && a.Type == AccountTypeAPIKey && a.IsExternalOpenAICompatible()\n}\n\n''',
)

# Preserve upstream's scoped fast-policy body while forwarding the incoming path/query.
replace_once(
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    '''\t\tupstreamBody = updatedBody\n\t}\n\tupstreamBody = updatedBody\n\n\t// Grok Composer''',
    '''\t\tupstreamBody = updatedBody\n\t}\n\n\t// Grok Composer''',
)
replace_once(
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    '''\ttargetURL, err := s.rawChatCompletionsURL(account)\n''',
    '''\ttargetURL, err := s.rawChatCompletionsURL(account, externalOpenAIIncomingPath(c))\n''',
)

# Make composite and dedicated external groups select the correct scheduler platform.
regex_replace_once(
    "backend/internal/handler/openai_gateway_handler.go",
    r'''func openAICompatibleRequestPlatform\(ctx context\.Context, apiKey \*service\.APIKey\) string \{.*?\n\}\n''',
    '''func openAICompatibleRequestPlatform(ctx context.Context, apiKey *service.APIKey) string {\n\tif platform, ok := service.ResolvedTargetPlatformFromContext(ctx); ok {\n\t\tnormalized := service.NormalizeOpenAICompatiblePlatformForRouting(platform)\n\t\tswitch normalized {\n\t\tcase service.PlatformGrok, service.PlatformExternalOpenAI:\n\t\t\treturn normalized\n\t\tdefault:\n\t\t\treturn service.PlatformOpenAI\n\t\t}\n\t}\n\tif apiKey != nil && apiKey.Group != nil {\n\t\tnormalized := service.NormalizeOpenAICompatiblePlatformForRouting(apiKey.Group.Platform)\n\t\tswitch normalized {\n\t\tcase service.PlatformGrok, service.PlatformExternalOpenAI:\n\t\t\treturn normalized\n\t\t}\n\t}\n\treturn service.PlatformOpenAI\n}\n''',
    flags=re.S,
)

replace_once(
    "backend/internal/handler/openai_gateway_count_tokens.go",
    '''\trequestPlatform := openAICompatibleRequestPlatform(apiKey)\n\tcurrentRoutingModel := routingModel\n''',
    '''\tcurrentRoutingModel := routingModel\n''',
)
replace_once(
    "backend/internal/handler/openai_embeddings.go",
    '''\trequestPlatform := openAICompatibleRequestPlatform(apiKey)\n''',
    '''\trequestPlatform := openAICompatibleRequestPlatform(c.Request.Context(), apiKey)\n''',
)
replace_once(
    "backend/internal/handler/openai_rerank.go",
    '''\trequestPlatform := openAICompatibleRequestPlatform(apiKey)\n''',
    '''\trequestPlatform := openAICompatibleRequestPlatform(c.Request.Context(), apiKey)\n''',
)
replace_once(
    "backend/internal/handler/openai_rerank.go",
    '''\t\t\tservice.OpenAIEndpointCapabilityRerank,\n\t\t\tfalse,\n\t\t\trequestPlatform,\n''',
    '''\t\t\tservice.OpenAIEndpointCapabilityRerank,\n\t\t\tfalse,\n\t\t\tfalse,\n\t\t\ttrue,\n\t\t\trequestPlatform,\n''',
)
replace_exact(
    "backend/internal/handler/openai_rerank.go",
    '''h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)''',
    '''h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, account.GetMappedModel(reqModel), false, nil)''',
    2,
)
replace_once(
    "backend/internal/handler/openai_rerank.go",
    '''h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)''',
    '''h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, account.GetMappedModel(reqModel), true, nil)''',
)

# Route embeddings and rerank through the external compatibility lane without widening OpenAI-only media endpoints.
replace_once(
    "backend/internal/server/routes/gateway.go",
    '''\tisOpenAIOnlyEndpointGatewayPlatform := func(c *gin.Context) bool {\n\t\treturn getGroupPlatform(c) == service.PlatformOpenAI\n\t}\n''',
    '''\tisEmbeddingsGatewayPlatform := func(c *gin.Context) bool {\n\t\tswitch service.NormalizeOpenAICompatiblePlatformForRouting(getGroupPlatform(c)) {\n\t\tcase service.PlatformOpenAI, service.PlatformExternalOpenAI:\n\t\t\treturn true\n\t\tdefault:\n\t\t\treturn false\n\t\t}\n\t}\n\tisRerankGatewayPlatform := func(c *gin.Context) bool {\n\t\treturn service.NormalizeOpenAICompatiblePlatformForRouting(getGroupPlatform(c)) == service.PlatformExternalOpenAI\n\t}\n\trerankHandler := func(c *gin.Context) {\n\t\tif isRerankGatewayPlatform(c) {\n\t\t\th.OpenAIGateway.Rerank(c)\n\t\t\treturn\n\t\t}\n\t\tservice.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)\n\t\tc.JSON(http.StatusNotFound, gin.H{\n\t\t\t"error": gin.H{\n\t\t\t\t"type":    "not_found_error",\n\t\t\t\t"message": "Rerank API is not supported for this platform",\n\t\t\t},\n\t\t})\n\t}\n''',
)
replace_once(
    "backend/internal/server/routes/gateway.go",
    '''\t\t\tif !isOpenAIOnlyEndpointGatewayPlatform(c) {\n''',
    '''\t\t\tif !isEmbeddingsGatewayPlatform(c) {\n''',
)

# Restore the external endpoint state declarations that were lost in the Vue file's structural conflict.
replace_once(
    "frontend/src/components/account/CreateAccountModal.vue",
    '''const apiKeyBaseUrl = ref('https://api.anthropic.com')\nconst apiKeyValue = ref('')\nconst upstreamBillingAutoProbeEnabled = ref(true)\n''',
    '''const apiKeyBaseUrl = ref('https://api.anthropic.com')\nconst apiKeyValue = ref('')\nconst externalOpenAIPreset = ref<'custom' | 'ark' | 'xunfei'>('custom')\nconst externalRequestPassthroughEnabled = ref(false)\nconst externalOpenAIEndpoints: { key: OpenAIEndpointCapability; label: string }[] = [\n  { key: 'chat_completions', label: 'Chat' },\n  { key: 'responses', label: 'Responses' },\n  { key: 'embeddings', label: 'Embeddings' },\n  { key: 'rerank', label: 'Rerank' }\n]\nconst defaultExternalEndpointPaths = {\n  chat_completions: '/chat/completions',\n  responses: '/responses',\n  embeddings: '/embeddings',\n  rerank: '/rerank'\n} satisfies Record<OpenAIEndpointCapability, string>\nconst externalEndpointBaseUrls = reactive<Record<OpenAIEndpointCapability, string>>({\n  chat_completions: '',\n  responses: '',\n  embeddings: '',\n  rerank: ''\n})\nconst externalEndpointPaths = reactive<Record<OpenAIEndpointCapability, string>>({ ...defaultExternalEndpointPaths })\nconst upstreamBillingAutoProbeEnabled = ref(true)\n''',
)

# Force the newly patched PostCSS release across Vue's transitive compiler dependency graph.
package_path = ROOT / "frontend/package.json"
package_data = json.loads(package_path.read_text(encoding="utf-8"))
overrides = package_data.setdefault("pnpm", {}).setdefault("overrides", {})
overrides["postcss@<8.5.12"] = ">=8.5.12"
package_path.write_text(json.dumps(package_data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

# Format the touched Go files before the workflow commits the port.
go_files = [
    "backend/internal/domain/constants.go",
    "backend/internal/service/domain_constants.go",
    "backend/internal/service/account.go",
    "backend/internal/service/openai_gateway_chat_completions_raw.go",
    "backend/internal/handler/openai_gateway_handler.go",
    "backend/internal/handler/openai_gateway_count_tokens.go",
    "backend/internal/handler/openai_embeddings.go",
    "backend/internal/handler/openai_rerank.go",
    "backend/internal/server/routes/gateway.go",
]
subprocess.run(["gofmt", "-w", *go_files], check=True)
