// Package ratelimit Tokenbucket based request rate limiter
package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/mailgun/timetools"
	"github.com/mailgun/ttlmap"
)

const DefaultKey = "CUMIRROR-TOKEN-LIMIT"

// TokenLimiter implements rate limiting middleware.
type OnlyTokenLimiter struct {
	defaultRates *RateSet
	bucketSets   *ttlmap.TtlMap
	capacity     int
	clock        timetools.TimeProvider
	mutex        sync.Mutex
}

// New constructs a `TokenLimiter` middleware instance.
func NewTokenLimiter(defaultRates *RateSet) (*OnlyTokenLimiter, error) {
	if defaultRates == nil || len(defaultRates.m) == 0 {
		return nil, fmt.Errorf("provide default rates")
	}

	tl := &OnlyTokenLimiter{
		defaultRates: defaultRates,
		clock:        &timetools.RealTime{},
		capacity:     DefaultCapacity,
	}

	bucketSets, err := ttlmap.NewMapWithProvider(tl.capacity, tl.clock)
	if err != nil {
		return nil, err
	}
	tl.bucketSets = bucketSets
	return tl, nil
}

func (tl *OnlyTokenLimiter) Consume(num int64, key string) (bool, error) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	if key == "" {
		key = DefaultKey
	}

	bucketSetI, exists := tl.bucketSets.Get(key)
	var bucketSet *TokenBucketSet

	if exists {
		bucketSet = bucketSetI.(*TokenBucketSet)
		bucketSet.Update(tl.defaultRates)
	} else {
		bucketSet = NewTokenBucketSet(tl.defaultRates, tl.clock)
		// We set ttl as 10 times rate period. E.g. if rate is 100 requests/second per client ip
		// the counters for this ip will expire after 10 seconds of inactivity
		tl.bucketSets.Set(key, bucketSet, int(bucketSet.maxPeriod/time.Second)*10+1)
	}
	delay, err := bucketSet.Consume(num)
	if err != nil || delay > 0 {
		return false, err
	}
	return true, err
}
