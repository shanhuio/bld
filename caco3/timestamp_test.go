package caco3

import (
	"testing"
	"time"
)

func TestTimestampRoundTrip(t *testing.T) {
	now := time.Now()
	got := newTimestamp(now).toTime().UnixNano()
	if got != now.UnixNano() {
		t.Errorf("roundtrip: %d != %d", got, now.UnixNano())
	}
}

func TestTimestampNilToTime(t *testing.T) {
	var ts *timestamp
	got := ts.toTime()
	var zero time.Time
	if !got.Equal(zero) {
		t.Errorf("nil timestamp toTime = %q, want zero", got)
	}
}

func TestReadTime(t *testing.T) {
	moment := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	got := readTime(func() time.Time { return moment })
	if !got.Equal(moment) {
		t.Errorf("readTime with fn: got %v, want %v", got, moment)
	}
	// Nil falls back to time.Now; just verify monotonicity.
	before := time.Now()
	after := readTime(nil)
	if after.Before(before) {
		t.Errorf("readTime(nil) returned a time before now")
	}
}
