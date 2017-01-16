package flags

import (
	"flag"
	"time"
)

type unixTime struct{ time time.Time }

func (t *unixTime) String() string {
	return t.time.Format(time.UnixDate)
}

func (t *unixTime) Set(s string) error {
	parsed, err := time.Parse(time.UnixDate, s)
	t.time = parsed
	return err
}

// UnixTime defines a flag for time.Time values formatted as Unix date.
// Call the returned function after flag.Parse to get the value.
func UnixTime(name string, value time.Time, usage string) func() time.Time {
	t := &unixTime{value}
	flag.Var(t, name, usage)
	return func() time.Time {
		return t.time
	}
}
