package generator

import (
	"context"
	"testing"
)

func TestNewNoopGenerator(t *testing.T) {
	gen := NewNoopGenerator()
	if gen == nil {
		t.Fatal("NewNoopGenerator returned nil")
	}
}

func TestNoopGenerator_Generate(t *testing.T) {
	gen := NewNoopGenerator()
	err := gen.Generate(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from noop generator, got nil")
	}
	if err.Error() != "project generator not yet implemented - coming in future phases" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNoopGenerator_Validate(t *testing.T) {
	gen := NewNoopGenerator()
	err := gen.Validate(nil)
	if err == nil {
		t.Error("Expected error from noop generator, got nil")
	}
	if err.Error() != "project generator validation not yet implemented - coming in future phases" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
