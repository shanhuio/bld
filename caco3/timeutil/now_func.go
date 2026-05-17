package timeutil

import (
	"time"
)

// ReadTime runs the function and returns the time if f is not null, or returns
// time.Now() if f is null.
func ReadTime(f func() time.Time) time.Time {
	if f == nil {
		return time.Now()
	}
	return f()
}
