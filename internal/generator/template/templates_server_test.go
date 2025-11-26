package template

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Server Template Tests

func TestServerTemplateRender(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "package http")
	assert.Contains(t, output, "type Server struct")
	assert.Contains(t, output, "func NewServer")
}

func TestServerValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "server.go", output, parser.AllErrors)
	assert.NoError(t, err, "Generated server.go should be valid Go code")
}

func TestServerStructDefinition(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "type Server struct", "should define Server struct")
	assert.Contains(t, output, "router chi.Router", "should have router field")
	assert.Contains(t, output, "config *config.ServerConfig", "should have config field")
	assert.Contains(t, output, "logger interfaces.Logger", "should have logger field")
	assert.Contains(t, output, "healthService interfaces.HealthService", "should have healthService field")
}

func TestServerConstructor(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func NewServer(cfg *config.ServerConfig, logger interfaces.Logger) *Server", "should have NewServer constructor with config and logger")
	assert.Contains(t, output, "return &Server{", "should return Server pointer")
	assert.Contains(t, output, "router: chi.NewRouter()", "should initialize chi router")
	assert.Contains(t, output, "config: cfg", "should store config")
	assert.Contains(t, output, "logger: logger", "should store logger")
}

func TestServerBuilderPattern(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func (s *Server) WithHealthService(svc interfaces.HealthService) *Server", "should have WithHealthService method")
	assert.Contains(t, output, "s.healthService = svc", "should set healthService field")
	assert.Contains(t, output, "return s", "should return self for chaining")
}

func TestServerRegisterRoutes(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func (s *Server) RegisterRoutes() *Server", "should have RegisterRoutes method")
	assert.Contains(t, output, "s.routes()", "should call routes method")
}

func TestServerGracefulShutdown(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func (s *Server) ListenAndServe() error", "should have ListenAndServe method")
	assert.Contains(t, output, "signal.Notify(quit, os.Interrupt)", "should handle shutdown signals cross-platform")
	assert.Contains(t, output, "srv.Shutdown(shutdownCtx)", "should call graceful shutdown")
}

func TestServerTimeouts(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "ReadTimeout:  s.config.ReadTimeout", "should use config read timeout")
	assert.Contains(t, output, "WriteTimeout: s.config.WriteTimeout", "should use config write timeout")
	assert.Contains(t, output, "IdleTimeout:  s.config.IdleTimeout", "should use config idle timeout")
	assert.Contains(t, output, "context.WithTimeout(context.Background(), s.config.ShutdownTimeout)", "should use config shutdown timeout")
}

func TestServerModuleNameInterpolation(t *testing.T) {
	testCases := []struct {
		name       string
		moduleName string
	}{
		{
			name:       "simple module",
			moduleName: "myapp",
		},
		{
			name:       "github module",
			moduleName: "github.com/user/project",
		},
		{
			name:       "nested module",
			moduleName: "example.com/org/team/service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewRenderer(templates.FS)
			data := TemplateData{
				ModuleName: tc.moduleName,
			}

			output, err := renderer.Render("internal/http/server.go.tmpl", data)
			require.NoError(t, err)

			expectedImport := tc.moduleName + "/internal/interfaces"
			assert.Contains(t, output, expectedImport, "should interpolate module name in import")
		})
	}
}

func TestServerBuilderChain(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func (s *Server) WithHealthService(svc interfaces.HealthService) *Server", "WithHealthService should return *Server")
	assert.Contains(t, output, "func (s *Server) RegisterRoutes() *Server", "RegisterRoutes should return *Server")
	assert.Contains(t, output, "return s", "Builder methods should return self for chaining")
}

func TestServerNoGlobalState(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, output, "var server", "should not use global server variable")
	assert.NotContains(t, output, "var config", "should not use global config variable")
	assert.NotContains(t, output, "var logger", "should not use global logger variable")
}

func TestServerErrorWrapping(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, `fmt.Errorf("failed to start server: %w", err)`)
}

func TestServerContextAwareLogging(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "s.logger.Info(ctx)")
	assert.Contains(t, output, "s.logger.Error(ctx)")
}

func TestServerTemplateRegistration(t *testing.T) {
	generatorFile := "../generator.go"
	content, err := os.ReadFile(generatorFile)
	require.NoError(t, err)

	generatorCode := string(content)
	assert.Contains(t, generatorCode, `"internal/http/server.go.tmpl"`, "server.go.tmpl should be registered in generator")
	assert.Contains(t, generatorCode, `"internal/http/server.go"`, "server.go output path should be registered")
}

// HTTP Routes Template Tests (internal/http/routes.go.tmpl)

func TestHTTPRoutesTemplateRender(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "package http")
	assert.Contains(t, output, "func (s *Server) routes()")
}

func TestHTTPRoutesValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "routes.go", output, parser.AllErrors)
	assert.NoError(t, err, "Generated routes.go should be valid Go code")
}

func TestHTTPRoutesMiddlewareOrder(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	requestIDIdx := strings.Index(output, "middleware.RequestID")
	compressIdx := strings.Index(output, "httpmiddleware.NewCompressMiddleware()")
	loggerIdx := strings.Index(output, "httpmiddleware.NewRequestLogger(s.logger)")
	realIPIdx := strings.Index(output, "middleware.RealIP")
	recovererIdx := strings.Index(output, "httpmiddleware.NewRecoverer(s.logger)")

	assert.Greater(t, compressIdx, requestIDIdx, "Compress should come after RequestID")
	assert.Greater(t, loggerIdx, compressIdx, "RequestLogger should come after Compress")
	assert.Greater(t, realIPIdx, loggerIdx, "RealIP should come after RequestLogger")
	assert.Greater(t, recovererIdx, realIPIdx, "Recoverer should come after RealIP")
}

func TestHTTPRoutesCompressMiddleware(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "httpmiddleware.NewCompressMiddleware()", "should use compress middleware")
}

func TestHTTPRoutesMarkerSections(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "// TRACKS:API_ROUTES:BEGIN", "should have API routes begin marker")
	assert.Contains(t, output, "// TRACKS:API_ROUTES:END", "should have API routes end marker")
	assert.Contains(t, output, "// TRACKS:WEB_ROUTES:BEGIN", "should have web routes begin marker")
	assert.Contains(t, output, "// TRACKS:WEB_ROUTES:END", "should have web routes end marker")
	assert.Contains(t, output, "// TRACKS:PROTECTED_ROUTES:BEGIN", "should have protected routes begin marker")
	assert.Contains(t, output, "// TRACKS:PROTECTED_ROUTES:END", "should have protected routes end marker")
}

func TestHTTPRoutesAPIHealthCheckRoute(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "s.router.Get(routes.APIHealth, s.handleHealthCheck())", "should use routes.APIHealth constant")
}

func TestHTTPRoutesWebRoutesEmpty(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	webBegin := strings.Index(output, "// TRACKS:WEB_ROUTES:BEGIN")
	webEnd := strings.Index(output, "// TRACKS:WEB_ROUTES:END")

	require.NotEqual(t, -1, webBegin, "should have web routes begin marker")
	require.NotEqual(t, -1, webEnd, "should have web routes end marker")

	section := output[webBegin:webEnd]
	assert.Contains(t, section, "s.router.Get(routes.Home, s.handleHome(counterHandler))", "should register home route with counter")
	assert.Contains(t, section, "s.router.Get(routes.About, s.handleAbout())", "should register about route")
}

func TestHTTPRoutesProtectedRoutesGroup(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "s.router.Group(func(r chi.Router) {", "should create protected routes group")
	assert.Contains(t, output, "// TRACKS:PROTECTED_ROUTES:BEGIN", "should have protected routes section inside group")
}

func TestHTTPRoutesHandleHealthCheckHelper(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func (s *Server) handleHealthCheck() nethttp.HandlerFunc", "should have handleHealthCheck helper")
	assert.Contains(t, output, "handler := handlers.NewHealthHandler(s.healthService, s.logger)", "should instantiate handler with service and logger")
	assert.Contains(t, output, "return handler.Check", "should return handler method")
}

func TestHTTPRoutesModuleNameInterpolation(t *testing.T) {
	testCases := []struct {
		name       string
		moduleName string
	}{
		{
			name:       "simple module",
			moduleName: "myapp",
		},
		{
			name:       "github module",
			moduleName: "github.com/user/project",
		},
		{
			name:       "nested module",
			moduleName: "example.com/org/team/service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewRenderer(templates.FS)
			data := TemplateData{
				ModuleName: tc.moduleName,
			}

			output, err := renderer.Render("internal/http/routes.go.tmpl", data)
			require.NoError(t, err)

			expectedHandlersImport := tc.moduleName + "/internal/http/handlers"
			expectedRoutesImport := tc.moduleName + "/internal/http/routes"

			assert.Contains(t, output, expectedHandlersImport, "should interpolate module name in handlers import")
			assert.Contains(t, output, expectedRoutesImport, "should interpolate module name in routes import")
		})
	}
}

func TestHTTPRoutesHashfsImport(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, `"github.com/benbjohnson/hashfs"`, "should import hashfs package")
}

func TestHTTPRoutesNetHTTPImport(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, `nethttp "net/http"`, "should import net/http with nethttp alias")
}

func TestHTTPRoutesAssetsImport(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, `"github.com/example/testapp/internal/assets"`, "should import assets package")
}

func TestHTTPRoutesHashfsFileServer(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/routes.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, `s.router.Handle("/assets/*"`, "should register /assets/* route")
	assert.Contains(t, output, `nethttp.StripPrefix("/assets/"`, "should strip /assets/ prefix using nethttp alias")
	assert.Contains(t, output, `hashfs.FileServer(assets.FileSystem())`, "should use hashfs.FileServer with assets.FileSystem()")
}

// Server Test Template Tests (internal/http/server_test.go.tmpl)

func TestServerTestTemplateRender(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "package http")
	assert.Contains(t, output, "func TestServer_NewServer")
}

func TestServerTestValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "server_test.go", output, parser.AllErrors)
	assert.NoError(t, err, "Generated server_test.go should be valid Go code")
}

func TestServerTestUsesHTTPTest(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "net/http/httptest", "should import httptest package")
	assert.Contains(t, output, "httptest.NewRequest", "should use httptest.NewRequest")
	assert.Contains(t, output, "httptest.NewRecorder", "should use httptest.NewRecorder")
}

func TestServerTestUsesMocks(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)

	mocksImport := data.ModuleName + "/tests/mocks"
	loggingImport := data.ModuleName + "/internal/logging"
	assert.Contains(t, output, mocksImport, "should import from tests/mocks")
	assert.Contains(t, output, loggingImport, "should import from internal/logging")
	assert.Contains(t, output, "newTestLogger()", "should use test logger helper")
	assert.Contains(t, output, "mocks.NewMockHealthService", "should use MockHealthService")
}

func TestServerTestHasIntegrationTests(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func TestServer_NewServer(t *testing.T)")
	assert.Contains(t, output, "func TestServer_HealthEndpoint(t *testing.T)")
	assert.Contains(t, output, "func TestServer_NotFoundRoute(t *testing.T)")
}

func TestServerTestUsesRealRouter(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "srv.router.ServeHTTP(rr, req)", "should use real chi.Router via ServeHTTP")
	assert.NotContains(t, output, "MockRouter", "should not mock chi.Router")
}

func TestServerTestModuleNameInterpolation(t *testing.T) {
	testCases := []struct {
		name       string
		moduleName string
	}{
		{
			name:       "simple module",
			moduleName: "myapp",
		},
		{
			name:       "github module",
			moduleName: "github.com/user/project",
		},
		{
			name:       "nested module",
			moduleName: "example.com/org/team/service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := NewRenderer(templates.FS)
			data := TemplateData{
				ModuleName: tc.moduleName,
			}

			output, err := renderer.Render("internal/http/server_test.go.tmpl", data)
			require.NoError(t, err)

			expectedConfigImport := tc.moduleName + "/internal/config"
			expectedMocksImport := tc.moduleName + "/tests/mocks"

			assert.Contains(t, output, expectedConfigImport, "should interpolate module name in config import")
			assert.Contains(t, output, expectedMocksImport, "should interpolate module name in mocks import")
		})
	}
}
