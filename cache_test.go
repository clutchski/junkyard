package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheSize(t *testing.T) {
	c := newCache(3, time.Minute)
	assert := assert.New(t)
	assert.Equal(c.Len(), 0)

	b := []byte{1, 2, 3}

	c.Add("1", b)
	c.Add("2", b)
	c.Add("3", b)
	assert.Equal(c.Len(), 3)
	assert.Equal(c.Get("3"), b)
	c.Add("4", b)
	assert.Equal(c.Len(), 3)
	assert.Equal(c.Get("4"), b)
}

func TestCacheExpiry(t *testing.T) {
	c := newCache(10, time.Minute)
	assert := assert.New(t)

	now := time.Now().UTC()
	b := []byte{1, 2, 3}
	c.Add("x1", b)
	assert.Equal(c.Get("x1"), b)

	c.addAt("x2", b, now.Add(-time.Minute-time.Second))
	assert.Empty(c.Get("x2"))
}
