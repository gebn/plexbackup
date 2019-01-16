// Package stamp provides build-time workspace status information at runtime,
// including the code version and build metadata.
package stamp

import (
	"fmt"

	"github.com/gebn/plexbackup/stamp/build"
	"github.com/gebn/plexbackup/stamp/code"
)

// String returns a human-readable summary of the code version and build
// metadata.
func String() string {
	return fmt.Sprintf("%v, built by %v", code.String(), build.String())
}
