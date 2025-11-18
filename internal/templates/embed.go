// Package templates provides embedded template files for project generation.
// All template files are embedded at build time using Go's embed package,
// eliminating the need for external file dependencies at runtime.
package templates

import "embed"

// FS contains all embedded template files from the project and examples directories.
// The all:project pattern embeds all files recursively from the project directory.
// The all:examples pattern embeds reference templates used for future code generation.
//
// The embedded filesystem uses forward slashes (/) as path separators
// regardless of the host operating system, ensuring cross-platform compatibility.
//
//go:embed all:project all:examples
var FS embed.FS
