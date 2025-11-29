package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentifierTemplateRenders(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName:  "github.com/test/app",
		ProjectName: "app",
	}

	tmplFiles := []string{
		"internal/pkg/identifier/uuid.go.tmpl",
		"internal/pkg/identifier/uuid_test.go.tmpl",
	}

	for _, tmpl := range tmplFiles {
		t.Run(tmpl, func(t *testing.T) {
			result, err := renderer.Render(tmpl, data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "package identifier")
		})
	}
}

func TestIdentifierTemplateContainsFunctions(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{}

	result, err := renderer.Render("internal/pkg/identifier/uuid.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func NewID()")
	assert.Contains(t, result, "func ValidateID(")
	assert.Contains(t, result, "func ExtractTimestamp(")
	assert.Contains(t, result, "github.com/gofrs/uuid/v5")
}
