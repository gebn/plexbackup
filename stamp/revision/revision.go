// Package revision embeds information about the revision contained
// in this library, including the commit identifier, branch name, and
// description of the version (e.g. tag name).
package revision

import (
	"fmt"
)

var (
	// The SHA-1 hash of the commit that was built to produce this
	// library.
	Commit string

	// The name of the branch the above commit was on at the time it
	// was built. The CI pipeline checks out a specific commit,
	// meaning Git is in a detached HEAD state, so this will always
	// be "HEAD" for official releases.
	Branch string

	// The output of `git describe --always --tags --dirty`. For
	// official releases, this will be the tag name, which follows
	// semver, e.g. `v1.0.0`.
	Describe string
)

// String returns a human-readable summary of the revision metadata.
func String() string {
	return fmt.Sprintf("%v (%v, %v)", Describe, Commit, Branch)
}
