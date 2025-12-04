package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	logger := zerolog.Nop()
	return logger.WithContext(context.Background())
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name   string
		driver string
	}{
		{name: "sqlite3", driver: "sqlite3"},
		{name: "postgres", driver: "postgres"},
		{name: "go-libsql", driver: "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.driver)

			require.NotNil(t, m)
			assert.Equal(t, tt.driver, m.driver)
			assert.Empty(t, m.databaseURL)
			assert.Nil(t, m.db)
			assert.False(t, m.envLoaded)
		})
	}
}

func TestManager_GetDriver(t *testing.T) {
	m := NewManager("postgres")
	assert.Equal(t, "postgres", m.GetDriver())
}

func TestManager_GetDatabaseURL_BeforeLoad(t *testing.T) {
	m := NewManager("postgres")
	assert.Empty(t, m.GetDatabaseURL())
}

func TestManager_LoadEnv_NoEnvFile(t *testing.T) {
	ctx := testContext()
	m := NewManager("postgres")

	tmpDir := t.TempDir()
	err := m.LoadEnv(ctx, tmpDir)

	require.NoError(t, err)
	assert.True(t, m.envLoaded)
}

func TestManager_LoadEnv_WithEnvFile(t *testing.T) {
	ctx := testContext()
	m := NewManager("postgres")

	tmpDir := t.TempDir()
	envContent := "DATABASE_URL=postgres://localhost/testdb\n"
	envPath := filepath.Join(tmpDir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0600))

	os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("DATABASE_URL")

	err := m.LoadEnv(ctx, tmpDir)

	require.NoError(t, err)
	assert.True(t, m.envLoaded)
	assert.Equal(t, "postgres://localhost/testdb", m.GetDatabaseURL())
}

func TestManager_LoadEnv_EnvVarTakesPrecedence(t *testing.T) {
	ctx := testContext()
	m := NewManager("postgres")

	tmpDir := t.TempDir()
	envContent := "DATABASE_URL=postgres://localhost/from_file\n"
	envPath := filepath.Join(tmpDir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0600))

	os.Setenv("DATABASE_URL", "postgres://localhost/from_env")
	defer os.Unsetenv("DATABASE_URL")

	err := m.LoadEnv(ctx, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "postgres://localhost/from_env", m.GetDatabaseURL())
}

func TestManager_Connect_BeforeLoadEnv(t *testing.T) {
	ctx := testContext()
	m := NewManager("postgres")

	_, err := m.Connect(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEnvNotLoaded)
}

func TestManager_Connect_NoDatabaseURL(t *testing.T) {
	ctx := testContext()
	m := NewManager("postgres")

	tmpDir := t.TempDir()
	os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("DATABASE_URL")

	require.NoError(t, m.LoadEnv(ctx, tmpDir))

	_, err := m.Connect(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDatabaseURLNotSet)
}

func TestManager_Close_NotConnected(t *testing.T) {
	m := NewManager("postgres")

	err := m.Close()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotConnected)
}

func TestManager_IsConnected(t *testing.T) {
	m := NewManager("postgres")

	assert.False(t, m.IsConnected())
}

func TestManager_sqlDriverName(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
		wantErr  bool
	}{
		{driver: "postgres", expected: "postgres", wantErr: false},
		{driver: "sqlite3", expected: "", wantErr: true},
		{driver: "go-libsql", expected: "", wantErr: true},
		{driver: "unknown", expected: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			m := NewManager(tt.driver)
			name, err := m.sqlDriverName()

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrUnsupportedDriver)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, name)
			}
		})
	}
}
