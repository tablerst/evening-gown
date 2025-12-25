package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// PublicCache provides best-effort caching for public website APIs.
//
// Design goals:
// - Safe defaults: if Redis is disabled/unavailable, handlers must keep working.
// - "Next read is consistent": admin writes bump a version key, and public cache keys
//   include that version so old entries are automatically bypassed.
// - Minimal coupling: store final JSON response bytes for hot endpoints.
//
// NOTE: This cache is intentionally simple and does not implement stampede control
// (singleflight/locks) yet because the project is currently single-instance.
// It can be added later if needed.

type PublicCache struct {
	rdb *redis.Client
}

const (
	publicProductsVerKey = "eg:public:ver:products"
	publicUpdatesVerKey  = "eg:public:ver:updates"

	notFoundMarker = "__NOT_FOUND__"
)

func NewPublicCache(rdb *redis.Client) *PublicCache {
	return &PublicCache{rdb: rdb}
}

func (c *PublicCache) enabled() bool {
	return c != nil && c.rdb != nil
}

func (c *PublicCache) ProductsVersion(ctx context.Context) int64 {
	return c.getVersion(ctx, publicProductsVerKey)
}

func (c *PublicCache) UpdatesVersion(ctx context.Context) int64 {
	return c.getVersion(ctx, publicUpdatesVerKey)
}

func (c *PublicCache) BumpProductsVersion(ctx context.Context) (int64, error) {
	if !c.enabled() {
		return 0, nil
	}
	return c.rdb.Incr(ctx, publicProductsVerKey).Result()
}

func (c *PublicCache) BumpUpdatesVersion(ctx context.Context) (int64, error) {
	if !c.enabled() {
		return 0, nil
	}
	return c.rdb.Incr(ctx, publicUpdatesVerKey).Result()
}

func (c *PublicCache) GetJSONBytes(ctx context.Context, key string) ([]byte, bool, bool) {
	// returns (bytes, hit, isNotFoundMarker)
	if !c.enabled() {
		return nil, false, false
	}
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, false
		}
		return nil, false, false
	}
	if string(b) == notFoundMarker {
		return nil, true, true
	}
	return b, true, false
}

func (c *PublicCache) SetJSONBytes(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if !c.enabled() {
		return
	}
	if ttl <= 0 {
		return
	}
	_ = c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *PublicCache) SetNotFound(ctx context.Context, key string, ttl time.Duration) {
	if !c.enabled() {
		return
	}
	if ttl <= 0 {
		return
	}
	_ = c.rdb.Set(ctx, key, notFoundMarker, ttl).Err()
}

func (c *PublicCache) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	c.SetJSONBytes(ctx, key, b, ttl)
}

func (c *PublicCache) ProductsListKey(ver int64, season, category, availability, isNew string, limit, offset int) string {
	// Keep key stable by normalizing optional params.
	season = strings.TrimSpace(season)
	category = strings.TrimSpace(category)
	availability = strings.TrimSpace(availability)
	isNew = strings.TrimSpace(isNew)
	if isNew != "true" && isNew != "false" {
		isNew = ""
	}

	// Use a simple query-like format to keep it debuggable.
	return fmt.Sprintf("eg:public:products:list:v%d:season=%s:category=%s:availability=%s:is_new=%s:limit=%d:offset=%d", ver, escapeKeyPart(season), escapeKeyPart(category), escapeKeyPart(availability), isNew, limit, offset)
}

func (c *PublicCache) ProductDetailKey(ver int64, id uint) string {
	return fmt.Sprintf("eg:public:products:get:v%d:id=%d", ver, id)
}

func (c *PublicCache) UpdatesListKey(ver int64, limit, offset int) string {
	return fmt.Sprintf("eg:public:updates:list:v%d:limit=%d:offset=%d", ver, limit, offset)
}

func (c *PublicCache) UpdateDetailKey(ver int64, id uint) string {
	return fmt.Sprintf("eg:public:updates:get:v%d:id=%d", ver, id)
}

func (c *PublicCache) AssetAllowKey(productsVer int64, objectKey string) string {
	objectKey = strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	return fmt.Sprintf("eg:public:assets:allow:v%d:key=%s", productsVer, escapeKeyPart(objectKey))
}

func (c *PublicCache) BoolFromCache(ctx context.Context, key string) (val bool, hit bool) {
	if !c.enabled() {
		return false, false
	}
	raw, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, false
	}
	raw = strings.TrimSpace(raw)
	if raw == "1" || strings.EqualFold(raw, "true") {
		return true, true
	}
	if raw == "0" || strings.EqualFold(raw, "false") {
		return false, true
	}
	return false, false
}

func (c *PublicCache) SetBool(ctx context.Context, key string, v bool, ttl time.Duration) {
	if !c.enabled() {
		return
	}
	if ttl <= 0 {
		return
	}
	val := "0"
	if v {
		val = "1"
	}
	_ = c.rdb.Set(ctx, key, val, ttl).Err()
}

func (c *PublicCache) getVersion(ctx context.Context, key string) int64 {
	// Default version is 0.
	// If the version key doesn't exist, reads will use v0 keys.
	// The first bump (INCR on missing key) creates v1, which cleanly bypasses v0.
	if !c.enabled() {
		return 0
	}
	raw, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	if n < 0 {
		return 0
	}
	return n
}

func TTLWithKeyJitter(base time.Duration, key string, maxJitterFraction float64) time.Duration {
	if base <= 0 {
		return base
	}
	if maxJitterFraction <= 0 {
		return base
	}
	if maxJitterFraction > 1 {
		maxJitterFraction = 1
	}
	maxJitter := time.Duration(float64(base) * maxJitterFraction)
	if maxJitter <= 0 {
		return base
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	j := time.Duration(h.Sum32()%uint32(maxJitter))
	return base + j
}

func escapeKeyPart(s string) string {
	// Keep Redis keys readable while avoiding accidental separators.
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\n", "_")
	s = strings.ReplaceAll(s, "\r", "_")
	return s
}
