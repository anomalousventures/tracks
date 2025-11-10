//go:build docker

package integration

import (
	"testing"

)

func TestE2E_GoLibsql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "go-libsql")
}

func TestE2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E integration test in short mode")
	}
	runE2ETest(t, "postgres")
}

func TestDockerE2E_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "postgres")
}

func TestDockerE2E_GoLibsql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "go-libsql")
}

func TestDockerE2E_SQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker E2E integration test in short mode")
	}
	runDockerE2ETest(t, "sqlite3")
}
