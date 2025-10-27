package template

import "testing"

// TestTemplateDataStruct tests that TemplateData can be initialized and accessed
func TestTemplateDataStruct(t *testing.T) {
	data := TemplateData{
		ModuleName:  "github.com/user/myapp",
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		GoVersion:   "1.25",
		Year:        2025,
	}

	if data.ModuleName != "github.com/user/myapp" {
		t.Errorf("ModuleName = %q, want %q", data.ModuleName, "github.com/user/myapp")
	}

	if data.ProjectName != "myapp" {
		t.Errorf("ProjectName = %q, want %q", data.ProjectName, "myapp")
	}

	if data.DBDriver != "sqlite3" {
		t.Errorf("DBDriver = %q, want %q", data.DBDriver, "sqlite3")
	}

	if data.GoVersion != "1.25" {
		t.Errorf("GoVersion = %q, want %q", data.GoVersion, "1.25")
	}

	if data.Year != 2025 {
		t.Errorf("Year = %d, want %d", data.Year, 2025)
	}
}

// TestTemplateDataZeroValue tests the zero value of TemplateData
func TestTemplateDataZeroValue(t *testing.T) {
	var data TemplateData

	if data.ModuleName != "" {
		t.Errorf("zero value ModuleName = %q, want empty string", data.ModuleName)
	}

	if data.ProjectName != "" {
		t.Errorf("zero value ProjectName = %q, want empty string", data.ProjectName)
	}

	if data.DBDriver != "" {
		t.Errorf("zero value DBDriver = %q, want empty string", data.DBDriver)
	}

	if data.GoVersion != "" {
		t.Errorf("zero value GoVersion = %q, want empty string", data.GoVersion)
	}

	if data.Year != 0 {
		t.Errorf("zero value Year = %d, want 0", data.Year)
	}
}

// TestTemplateDataPartialInitialization tests partial struct initialization
func TestTemplateDataPartialInitialization(t *testing.T) {
	data := TemplateData{
		ModuleName: "github.com/user/app",
		Year:       2025,
	}

	if data.ModuleName != "github.com/user/app" {
		t.Errorf("ModuleName = %q, want %q", data.ModuleName, "github.com/user/app")
	}

	if data.Year != 2025 {
		t.Errorf("Year = %d, want %d", data.Year, 2025)
	}

	if data.ProjectName != "" {
		t.Errorf("unset ProjectName = %q, want empty string", data.ProjectName)
	}

	if data.DBDriver != "" {
		t.Errorf("unset DBDriver = %q, want empty string", data.DBDriver)
	}

	if data.GoVersion != "" {
		t.Errorf("unset GoVersion = %q, want empty string", data.GoVersion)
	}
}
