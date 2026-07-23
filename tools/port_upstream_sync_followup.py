#!/usr/bin/env python3
from pathlib import Path

ROOT = Path.cwd()


def replace_once(rel: str, old: str, new: str) -> None:
    path = ROOT / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{rel}: expected exactly one match, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")


replace_once(
    "backend/internal/server/routes/gateway.go",
    '''\tisEmbeddingsGatewayPlatform := func(c *gin.Context) bool {\n''',
    '''\tisOpenAIOnlyEndpointGatewayPlatform := func(c *gin.Context) bool {\n\t\treturn getGroupPlatform(c) == service.PlatformOpenAI\n\t}\n\tisEmbeddingsGatewayPlatform := func(c *gin.Context) bool {\n''',
)

replace_once(
    "frontend/src/types/index.ts",
    "export type GroupPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok' | 'composite'",
    "export type GroupPlatform = 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok' | 'composite' | 'external_openai_compatible' | 'volcengine_coding' | 'xunfei_coding'",
)
