// Package stamp provides build-time workspace status information at runtime,
// including the revision version and build metadata.
package stamp

import (
	"fmt"

	"github.com/gebn/plexbackup/stamp/build"
	"github.com/gebn/plexbackup/stamp/revision"
)

// String returns a human-readable summary of the revision version and build
// metadata.
func String() string {
	return fmt.Sprintf("%v, %v", revision.String(), build.String())
}
