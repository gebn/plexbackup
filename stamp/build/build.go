package build

import (
	"fmt"
	"strconv"
	"time"
)

var (
	User      string
	Host      string
	Timestamp string // seconds since epoch
	// TODO Go string
	// TODO Compiler string
)

func Time() time.Time {
	i, err := strconv.ParseInt(Timestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(i, 0)
}

func String() string {
	return fmt.Sprintf("%v@%v on %v", User, Host, Time().Format(time.RFC3339))
}
