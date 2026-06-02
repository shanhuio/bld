package lets

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func newTestBuildCache(t *testing.T) *buildCache {
	t.Helper()
	f := filepath.Join(t.TempDir(), "cache.sqlite")
	c, err := newBuildCache(f)
	if err != nil {
		t.Fatalf("newBuildCache: %v", err)
	}
	t.Cleanup(func() {
		if err := c.cache.db.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	return c
}

// fixedClock returns a clock function that always reports t.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func sampleBuilt() *built {
	return &built{
		Outs: []*fileStat{{
			Name:         "out.bin",
			Type:         fileTypeOut,
			Size:         42,
			ModTimestamp: 1000,
			Mode:         0644,
		}},
		Dockers: []*dockerSum{{
			Repo: "repo", Tag: "tag", ID: "abc123",
		}},
	}
}

func TestBuildCache_putGet(t *testing.T) {
	c := newTestBuildCache(t)
	in := sampleBuilt()
	if err := c.put("k1", in); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := c.get("k1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Errorf("get returned %+v, want %+v", got, in)
	}
}

func TestBuildCache_getMissing(t *testing.T) {
	c := newTestBuildCache(t)
	_, err := c.get("nope")
	if !errors.Is(err, errCacheMiss) {
		t.Errorf("get missing: got %v, want errCacheMiss", err)
	}
}

func TestBuildCache_removeMissingIsNoop(t *testing.T) {
	c := newTestBuildCache(t)
	if err := c.remove("nope"); err != nil {
		t.Errorf("remove missing: got %v, want nil", err)
	}
}

func TestBuildCache_removeThenGetMisses(t *testing.T) {
	c := newTestBuildCache(t)
	if err := c.put("k", sampleBuilt()); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := c.remove("k"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := c.get("k"); !errors.Is(err, errCacheMiss) {
		t.Errorf("get after remove: got %v, want errCacheMiss", err)
	}
}

func TestBuildCache_expired(t *testing.T) {
	c := newTestBuildCache(t)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c.clock = fixedClock(base)

	if err := c.put("k", sampleBuilt()); err != nil {
		t.Fatalf("put: %v", err)
	}

	// Right at expiry: now == createTime + expire; Before is false => miss.
	c.clock = fixedClock(base.Add(c.expire))
	if _, err := c.get("k"); !errors.Is(err, errCacheMiss) {
		t.Errorf("at expiry: got %v, want errCacheMiss", err)
	}

	// Past expiry.
	c.clock = fixedClock(base.Add(c.expire + time.Second))
	if _, err := c.get("k"); !errors.Is(err, errCacheMiss) {
		t.Errorf("past expiry: got %v, want errCacheMiss", err)
	}
}

func TestBuildCache_notYetExpired(t *testing.T) {
	c := newTestBuildCache(t)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c.clock = fixedClock(base)

	in := sampleBuilt()
	if err := c.put("k", in); err != nil {
		t.Fatalf("put: %v", err)
	}

	c.clock = fixedClock(base.Add(c.expire - time.Second))
	got, err := c.get("k")
	if err != nil {
		t.Fatalf("just before expiry: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Errorf("got %+v, want %+v", got, in)
	}
}

func TestBuildCache_putOverwrites(t *testing.T) {
	c := newTestBuildCache(t)

	first := &built{Outs: []*fileStat{{Name: "a"}}}
	second := &built{Outs: []*fileStat{{Name: "b"}}}

	if err := c.put("k", first); err != nil {
		t.Fatalf("first put: %v", err)
	}
	if err := c.put("k", second); err != nil {
		t.Fatalf("second put: %v", err)
	}
	got, err := c.get("k")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Outs) != 1 || got.Outs[0].Name != "b" {
		t.Errorf("got %+v, want second value", got)
	}
}
