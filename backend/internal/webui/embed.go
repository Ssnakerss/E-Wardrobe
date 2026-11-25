// Package webui embeds the production web build into the Go binary.
// The dist/ directory is populated by the web build step.
package webui

import "embed"

//go:embed dist
var Dist embed.FS
