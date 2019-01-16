// Package build embeds information about when the library was built,
// by whom, and on what machine.
package build

import (
	"fmt"
	"strconv"
	"time"
)

var (
	// User is the username that built the library.
	User string

	// Host is the unqualified hostname of the machine the library was
	// built on.
	Host string

	// Timestamp is when the library was built, represented as the
	// number of seconds since the epoch. Use Time() to get this value
	// is a form that is easier to work with.
	Timestamp string

	// TODO Go string
	// TODO Compiler string
)

// Time returns the Timestamp string in its native representation.
func Time() time.Time {
	i, err := strconv.ParseInt(Timestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(i, 0)
}

// String returns a human-readable summary of the build metadata.
func String() string {
	return fmt.Sprintf("%v@%v on %v", User, Host, Time().Format(time.RFC3339))
}
