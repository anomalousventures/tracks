package generator

import (
	"context"
	"fmt"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
)

// ProjectConfig and other generator types remain in this package.
// The ProjectGenerator interface has moved to internal/cli/interfaces/generator.go
// following ADR-002 (interfaces defined by consumers).

type noopGenerator struct{}

// NewNoopGenerator returns a placeholder generator that returns "not yet implemented" errors.
// This is used until the actual generator is implemented in future phases.
func NewNoopGenerator() interfaces.ProjectGenerator {
	return &noopGenerator{}
}

func (n *noopGenerator) Generate(ctx context.Context, cfg any) error {
	return fmt.Errorf("project generator not yet implemented - coming in future phases")
}

func (n *noopGenerator) Validate(cfg any) error {
	return fmt.Errorf("project generator validation not yet implemented - coming in future phases")
}
