package testutil

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertValidGoCode(t *testing.T, code, filename string) {
	t.Helper()

	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filename, code, parser.AllErrors)
	require.NoError(t, err)
}

func AssertContainsAll(t *testing.T, code string, items []string) {
	t.Helper()

	for _, item := range items {
		assert.Contains(t, code, item)
	}
}
