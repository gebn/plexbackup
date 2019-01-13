package code

import (
	"fmt"
)

var (
	Commit   string
	Branch   string
	Describe string
)

func String() string {
	return fmt.Sprintf("%v (%v, %v)", Describe, Commit, Branch)
}
