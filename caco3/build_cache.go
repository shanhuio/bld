package caco3

import (
	"errors"
	"fmt"
	"time"

	"shanhu.io/bld/caco3/timeutil"
)

type buildCache struct {
	cache  *kvTable
	expire time.Duration
	clock  func() time.Time
}

func newBuildCache(f string) (*buildCache, error) {
	cache, err := openKVTable(f)
	if err != nil {
		return nil, fmt.Errorf("open cache table: %w", err)
	}

	return &buildCache{
		expire: time.Hour * 24 * 7,
		cache:  cache,
	}, nil
}

type buildCacheEntry struct {
	Key        string              `json:"K"`
	CreateTime *timeutil.Timestamp `json:"T"`
	Built      *built              `json:"B"`
}

func (c *buildCache) put(k string, out *built) error {
	t := timeutil.ReadTime(c.clock)
	entry := &buildCacheEntry{
		Key:        k,
		Built:      out,
		CreateTime: timeutil.NewTimestamp(t),
	}
	return c.cache.replace(k, entry)
}

var errCacheMiss = errors.New("cache miss")

func (c *buildCache) get(k string) (*built, error) {
	entry := new(buildCacheEntry)
	if err := c.cache.get(k, entry); err != nil {
		if errors.Is(err, errKeyNotFound) {
			return nil, errCacheMiss
		}
		return nil, fmt.Errorf("get from cache: %w", err)
	}

	now := timeutil.ReadTime(c.clock)
	expire := timeutil.Time(entry.CreateTime).Add(c.expire)
	if now.Before(expire) {
		return entry.Built, nil
	}
	return nil, errCacheMiss
}

func (c *buildCache) remove(k string) error {
	if err := c.cache.remove(k); err != nil {
		if errors.Is(err, errKeyNotFound) {
			return nil
		}
		return err
	}
	return nil
}
