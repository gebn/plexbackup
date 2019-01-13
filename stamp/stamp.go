package stamp

import (
	"fmt"

	"github.com/gebn/plexbackup/stamp/build"
	"github.com/gebn/plexbackup/stamp/code"
)

func String() string {
	return fmt.Sprintf("%v, built by %v", code.String(), build.String())
}
