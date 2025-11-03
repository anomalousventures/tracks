package template

import (
	"go/parser"
	"go/token"
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

func TestNewConfig(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func NewConfig() *Config", "should have NewConfig constructor")
	assert.Contains(t, output, "Port:            \":8080\"", "should have default port")
	assert.Contains(t, output, "ReadTimeout:     15 * time.Second", "should have default read timeout")
	assert.Contains(t, output, "WriteTimeout:    15 * time.Second", "should have default write timeout")
	assert.Contains(t, output, "IdleTimeout:     60 * time.Second", "should have default idle timeout")
	assert.Contains(t, output, "ShutdownTimeout: 30 * time.Second", "should have default shutdown timeout")
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
	assert.Contains(t, output, "config *Config", "should have config field")
	assert.Contains(t, output, "healthService interfaces.HealthService", "should have healthService field")
}

func TestServerConfigStructDefinition(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "type Config struct", "should define Config struct")
	assert.Contains(t, output, "Port            string", "should have Port field")
	assert.Contains(t, output, "ReadTimeout     time.Duration", "should have ReadTimeout field")
	assert.Contains(t, output, "WriteTimeout    time.Duration", "should have WriteTimeout field")
	assert.Contains(t, output, "IdleTimeout     time.Duration", "should have IdleTimeout field")
	assert.Contains(t, output, "ShutdownTimeout time.Duration", "should have ShutdownTimeout field")
}

func TestServerConstructor(t *testing.T) {
	renderer := NewRenderer(templates.FS)
	data := TemplateData{
		ModuleName: "github.com/example/testapp",
	}

	output, err := renderer.Render("internal/http/server.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, output, "func NewServer(cfg *Config) *Server", "should have NewServer constructor")
	assert.Contains(t, output, "return &Server{", "should return Server pointer")
	assert.Contains(t, output, "router: chi.NewRouter()", "should initialize chi router")
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
	assert.Contains(t, output, "signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)", "should handle shutdown signals")
	assert.Contains(t, output, "srv.Shutdown(ctx)", "should call graceful shutdown")
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
	realIPIdx := strings.Index(output, "middleware.RealIP")
	loggerIdx := strings.Index(output, "middleware.Logger")
	recovererIdx := strings.Index(output, "middleware.Recoverer")

	assert.Greater(t, realIPIdx, requestIDIdx, "RealIP should come after RequestID")
	assert.Greater(t, loggerIdx, realIPIdx, "Logger should come after RealIP")
	assert.Greater(t, recovererIdx, loggerIdx, "Recoverer should come after Logger")
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
	lines := strings.Split(section, "\n")
	assert.Len(t, lines, 2, "web routes section should only contain begin marker and empty line")
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

	assert.Contains(t, output, "func (s *Server) handleHealthCheck() http.HandlerFunc", "should have handleHealthCheck helper")
	assert.Contains(t, output, "handler := handlers.NewHealthHandler(s.healthService)", "should instantiate handler with DI")
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
