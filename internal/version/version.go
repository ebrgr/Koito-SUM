// Package version supplies the default version for local and release builds.
package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var source string

var Current = strings.TrimSpace(source)
