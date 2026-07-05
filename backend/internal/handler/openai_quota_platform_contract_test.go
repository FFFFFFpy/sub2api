package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIRecordUsageInputsCarryQuotaPlatform(t *testing.T) {
	files := []string{
		"openai_gateway_handler.go",
		"openai_chat_completions.go",
		"openai_embeddings.go",
		"openai_images.go",
	}

	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
			require.NoError(t, err)

			var missing []token.Position
			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.CompositeLit)
				if !ok || !isOpenAIRecordUsageInputLiteral(literal.Type) {
					return true
				}
				if !compositeLiteralHasKey(literal, "QuotaPlatform") {
					missing = append(missing, fset.Position(literal.Lbrace))
				}
				return true
			})

			require.Empty(t, missing, "OpenAI usage post-billing must receive request-time QuotaPlatform")
		})
	}
}

func TestOpenAIChatCompletionsNoAccountClassificationUsesRequestPlatform(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(".", "openai_chat_completions.go"), nil, 0)
	require.NoError(t, err)

	var badCalls []token.Position
	var requestPlatformCalls int
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isIdentCall(call.Fun, "classifyNoAccountErrorFromGin") {
			return true
		}
		if len(call.Args) < 6 {
			return true
		}
		if ident, ok := call.Args[5].(*ast.Ident); ok && ident.Name == "requestPlatform" {
			requestPlatformCalls++
			return true
		}
		badCalls = append(badCalls, fset.Position(call.Lparen))
		return true
	})

	require.Empty(t, badCalls, "ChatCompletions no-account classification must pass requestPlatform, not a hard-coded platform")
	require.Equal(t, 2, requestPlatformCalls, "ChatCompletions has two no-account classification sites")
}

func isOpenAIRecordUsageInputLiteral(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "service" && selector.Sel.Name == "OpenAIRecordUsageInput"
}

func compositeLiteralHasKey(literal *ast.CompositeLit, key string) bool {
	for _, elt := range literal.Elts {
		pair, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := pair.Key.(*ast.Ident)
		if ok && ident.Name == key {
			return true
		}
	}
	return false
}

func isIdentCall(expr ast.Expr, name string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == name
}
