package flags

import (
	"flag"
	"time"
)

type unixTime time.Time

func (t *unixTime) String() string {
	return (*time.Time)(t).Format(time.UnixDate)
}

func (t *unixTime) Set(s string) error {
	parsed, err := time.Parse(time.UnixDate, s)
	*t = unixTime(parsed)
	return err
}

// UnixTime defines a flag for time.Time values formatted as Unix date.
func UnixTime(name string, value time.Time, usage string) *time.Time {
	t := &value
	flag.Var((*unixTime)(t), name, usage)
	return t
}
