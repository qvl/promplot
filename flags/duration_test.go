package flags

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		text    string
		parsed  time.Duration
		invalid bool
	}{
		{text: "1h30.5m3s", parsed: time.Hour + 30*time.Minute + 33*time.Second},
		{text: "1h", parsed: time.Hour},
		{text: "1h2h", parsed: 3 * time.Hour},
		{text: "30m1h", parsed: time.Hour + 30*time.Minute},
		{text: "1d", parsed: 24 * time.Hour},
		{text: "0.50d", parsed: 12 * time.Hour},
		{text: ".5d", parsed: 12 * time.Hour},
		{text: "1d1h", parsed: 25 * time.Hour},
		{text: "1h1d", parsed: 25 * time.Hour},
		{text: "1h1d60m", parsed: 26 * time.Hour},
		{text: "", invalid: true},
		{text: "bla", invalid: true},
	}

	for i, tt := range tests {
		var u durationValue
		if err := u.Set(tt.text); err != nil {
			if !tt.invalid {
				t.Errorf("parsing '%s' failed unexpectedly: %v", tt.text, err)
			}
			continue
		}
		if tt.invalid {
			t.Errorf("parsing '%s' should have failed", tt.text)
			continue
		}
		if time.Duration(u) != tt.parsed {
			t.Errorf(`
%d.
Input:    %s
Expected: %v
Got       %v`, i, tt.text, tt.parsed, u)
		}
	}

}
