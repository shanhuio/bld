package lets

import "time"

// timestamp is a JSON-stable moment-in-time, broken into separate
// seconds and nanoseconds components.
type timestamp struct {
	Sec  int64
	Nano int64 `json:",omitempty"`
}

func secNano(nano int64) (int64, int64) {
	sec := nano / 1e9
	nano -= sec * 1e9
	if nano < 0 {
		nano += 1e9
		sec--
	}
	return sec, nano
}

// newTimestamp captures t into a timestamp.
func newTimestamp(t time.Time) *timestamp {
	sec, nano := secNano(t.UnixNano())
	return &timestamp{Sec: sec, Nano: nano}
}

// toTime converts the timestamp back to time.Time in UTC. Returns a zero
// time if t is nil.
func (t *timestamp) toTime() time.Time {
	if t == nil {
		var zero time.Time
		return zero
	}
	return time.Unix(t.Sec, t.Nano).UTC()
}

// readTime returns f() if f is non-nil, otherwise time.Now().
func readTime(f func() time.Time) time.Time {
	if f == nil {
		return time.Now()
	}
	return f()
}
