package flags

import (
	"flag"
	"regexp"
	"strconv"
	"time"
)

type durationValue time.Duration

var matchDay = regexp.MustCompile("[-+]?[0-9]*(\\.[0-9]*)?d")

func (d *durationValue) Set(s string) error {
	if is := matchDay.FindStringIndex(s); is != nil {
		days, err := strconv.ParseFloat(s[is[0]:is[1]-1], 64)
		if err != nil {
			return err
		}
		s = s[:is[0]] + s[is[1]:] + strconv.FormatFloat(days*24, 'f', 6, 64) + "h"
	}
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d *durationValue) String() string {
	return (*time.Duration)(d).String()
}

// Duration defines a flag for time.Duration values.
// It works the same way as flag.Duration except that it als parses "XXd" values as days.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	t := &value
	flag.Var((*durationValue)(t), name, usage)
	return t
}
