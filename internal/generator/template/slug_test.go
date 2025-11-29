package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugTemplateRenders(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName:  "github.com/test/app",
		ProjectName: "app",
	}

	tmplFiles := []string{
		"internal/pkg/slug/slug.go.tmpl",
		"internal/pkg/slug/slug_test.go.tmpl",
	}

	for _, tmpl := range tmplFiles {
		t.Run(tmpl, func(t *testing.T) {
			result, err := renderer.Render(tmpl, data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "package slug")
		})
	}
}

func TestSlugTemplateContainsFunctions(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{}

	result, err := renderer.Render("internal/pkg/slug/slug.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func Generate()")
	assert.Contains(t, result, "func GenerateShort()")
	assert.Contains(t, result, "func Sanitize(")
	assert.Contains(t, result, "func ValidateUsername(")
	assert.Contains(t, result, "github.com/jaevor/go-nanoid")
}
