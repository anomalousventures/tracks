package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetsEmbedTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAssetsEmbedPackage(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "package assets", "should be in assets package")
}

func TestAssetsEmbedHashfsImport(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, `"github.com/benbjohnson/hashfs"`, "should import hashfs")
}

func TestAssetsEmbedDirective(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "//go:embed all:dist", "should embed dist directory")
}

func TestAssetsEmbedHashfsInit(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "var fsys *hashfs.FS", "should declare hashfs.FS variable")
	assert.Contains(t, result, "func init()", "should use init function")
	assert.Contains(t, result, "hashfs.NewFS", "should wrap with hashfs.NewFS")
}

func TestAssetsEmbedFileSystem(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func FileSystem() *hashfs.FS", "should return *hashfs.FS")
	assert.Contains(t, result, "return fsys", "should return the hashfs filesystem")
}

func TestAssetsEmbedAssetURL(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func AssetURL(path string) string", "should have AssetURL function")
	assert.Contains(t, result, "fsys.HashName(path)", "should use HashName for hashing")
}

func TestAssetsEmbedConvenienceFunctions(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "testapp",
	}

	result, err := renderer.Render("internal/assets/embed.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func CSSURL() string", "should have CSSURL function")
	assert.Contains(t, result, "func JSURL() string", "should have JSURL function")
	assert.Contains(t, result, `AssetURL("css/app.css")`, "CSSURL should call AssetURL with css path")
	assert.Contains(t, result, `AssetURL("js/app.js")`, "JSURL should call AssetURL with js path")
}
