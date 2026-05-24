package caco3

import (
	"errors"
	"fmt"
	"time"
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
	Key        string     `json:"K"`
	CreateTime *timestamp `json:"T"`
	Built      *built     `json:"B"`
}

func (c *buildCache) put(k string, out *built) error {
	t := readTime(c.clock)
	entry := &buildCacheEntry{
		Key:        k,
		Built:      out,
		CreateTime: newTimestamp(t),
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

	now := readTime(c.clock)
	expire := entry.CreateTime.toTime().Add(c.expire)
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
