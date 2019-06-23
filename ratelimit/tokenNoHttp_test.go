package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We've hit the limit and were able to proceed on the next time run
func TestTokenHitLimit(t *testing.T) {
	rates := NewRateSet()
	err := rates.Add(time.Second, 1, 1)
	require.NoError(t, err)

	l, err := NewTokenLimiter(rates)
	require.NoError(t, err)

	re, err := l.Consume(1, "test")
	require.NoError(t, err)
	assert.Equal(t, true, re)

	// Next request from the same key hits rate limit
	re, err = l.Consume(1, "test")
	require.NoError(t, err)
	assert.Equal(t, false, re)

	// Second later, the request from this key will succeed
	time.Sleep(time.Second)
	re, err = l.Consume(1, "test")
	require.NoError(t, err)
	assert.Equal(t, true, re)

	// Next request from other key don't hits rate limit
	re, err = l.Consume(1, "test1")
	require.NoError(t, err)
	assert.Equal(t, true, re)

	// Next request from the same key hits rate limit
	re, err = l.Consume(1, "test1")
	require.NoError(t, err)
	assert.Equal(t, false, re)

	// Second later, the request from this key will succeed
	time.Sleep(time.Second)
	re, err = l.Consume(1, "test1")
	require.NoError(t, err)
	assert.Equal(t, true, re)
}
