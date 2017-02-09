package flags

import (
	"testing"
	"time"
)

func TestUnixTime(t *testing.T) {
	tests := []struct {
		text    string
		parsed  time.Time
		invalid bool
	}{
		{text: "Mon Feb 4 10:08:05 CET 2017", parsed: time.Date(2017, 2, 4, 10, 7, 5, 8, time.UTC)},
		{text: "Tue Feb 4 10:08:05 CET 2017", parsed: time.Date(2017, 2, 4, 10, 7, 5, 8, time.UTC)},
		{text: "", invalid: true},
		{text: "Mon Feb 4 10:08:5 CET 2017", invalid: true},
		{text: "Mon Feb 4 10:8:05 CET 2017", invalid: true},
		{text: "Mon Feb 04 10:8:05 CET 2017", invalid: true},
		{text: "Feb 4 10:08:05 CET 2017", invalid: true},
		{text: "Mon Feb 4 10:08:05 2017", invalid: true},
		{text: "4 10:08:05 CET 2017", invalid: true},
		{text: "Mon Feb 4 10:08:05 CET", invalid: true},
	}

	for i, tt := range tests {
		u := unixTime{}
		if err := u.Set(tt.text); err != nil {
			if !tt.invalid {
				t.Errorf("parsing %s failed unexpectedly: %v", tt.text, err)
			}
			continue
		}
		if tt.invalid {
			t.Errorf("parsing %s should have failed", tt.text)
			continue
		}
		if time.Time(u).Equal(tt.parsed) {
			t.Errorf(`
%d.
Input:    %s
Expected: %v
Got       %v`, i, tt.text, tt.parsed, u)
		}
	}

}
