package timeutil

import (
	"time"
)

func secNano(nano int64) (int64, int64) {
	sec := nano / 1e9
	nano -= sec * 1e9
	if nano < 0 {
		nano += 1e9
		sec--
	}
	return sec, nano
}

// Timestamp is a struct to record a UTC timestamp.
// It is designed to be directly usable in Javascript.
type Timestamp struct {
	Sec  int64
	Nano int64 `json:",omitempty"`
}

// Time returns the time of this timestamp in UTC.
func (t *Timestamp) Time() time.Time {
	return time.Unix(t.Sec, t.Nano).UTC()
}

// NewTimestamp creates a new timestamp from the given time.
func NewTimestamp(t time.Time) *Timestamp {
	sec, nano := secNano(t.UnixNano())
	return &Timestamp{
		Sec:  sec,
		Nano: nano,
	}
}

// Time converts timestamp to time.Time .
func Time(ts *Timestamp) time.Time {
	if ts == nil {
		var zero time.Time
		return zero
	}
	return ts.Time()
}
