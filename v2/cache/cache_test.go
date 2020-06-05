package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/proto/v2/common"
)

const (
	wait          = 1 * time.Millisecond
	TTL           = 5 * wait
	NoTTL         = 0
	cacheFuncKey  = "myFunc"
	cacheFuncKey2 = "myFunc2"
	userID        = "myUserId"
	actionName    = "myActionname"
)

var (
	resource  = common.Origin{Id: "myId", Type: "myType"}
	keyFields = []string{userID, actionName, resource.Id, resource.Type}
	key       = Key(cacheFuncKey, keyFields...)
	key2      = Key(cacheFuncKey2, keyFields...)
)

func makeCache(t *testing.T, ttl time.Duration) *Cache {
	cache, err := New(ttl, 200)
	cache.SetLogger(log.NewSampleLogger(time.Second, 0, 10))
	require.NoError(t, err)
	require.NotNil(t, cache)
	cache.Clear()

	return cache
}

func Test_MakeCacheKey(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%s;%s;%s;%s;%s", cacheFuncKey, userID, actionName, resource.Id, resource.Type), string(key))
}

func Test_CacheNew(t *testing.T) {
	_ = makeCache(t, TTL)
}

func Test_CacheNewNoTTL(t *testing.T) {
	_ = makeCache(t, NoTTL)
}

func Test_CacheSetNoTTL(t *testing.T) {
	c := makeCache(t, TTL)

	require.True(t, c.Set(key, 1))
	time.Sleep(wait)

	c.SetTTL(NoTTL)
	assert.False(t, c.Set(key2, 2))
	assert.False(t, c.Exist(key))
}

func Test_CacheSet(t *testing.T) {
	c := makeCache(t, TTL)

	assert.True(t, c.Set(key, 1))
}

func Test_CacheSetWithNoTTL(t *testing.T) {
	c := makeCache(t, NoTTL)

	assert.False(t, c.Set(key, 1))
}

func Test_CacheGet(t *testing.T) {
	c := makeCache(t, TTL)

	require.True(t, c.Set(key, 1))

	time.Sleep(wait)

	val, ok := c.Get(key)

	require.True(t, ok)
	assert.Equal(t, 1, val)
}

func Test_CacheGetNonExistentKey(t *testing.T) {
	c := makeCache(t, TTL)
	_, ok := c.Get(key)
	assert.False(t, ok)
}

func Test_CacheGetNoTTL(t *testing.T) {
	c := makeCache(t, NoTTL)

	require.False(t, c.Set(key, 1))

	_, ok := c.Get(key)

	assert.False(t, ok)
}

func Test_CacheExist(t *testing.T) {
	c := makeCache(t, TTL)

	require.True(t, c.Set(key, 1))

	time.Sleep(wait)

	assert.True(t, c.Exist(key))
}

func Test_CacheExistsFalse(t *testing.T) {
	c := makeCache(t, TTL)

	assert.False(t, c.Exist(key))
}

func Test_CacheExistFalseAfterExpire(t *testing.T) {
	c := makeCache(t, TTL)

	require.True(t, c.Set(key, 1))

	time.Sleep(TTL + wait)

	assert.False(t, c.Exist(key))
}

func Test_CacheExpire(t *testing.T) {
	c := makeCache(t, TTL)

	assert.True(t, c.Set(key, 1))
	time.Sleep(wait)

	val, ok := c.Get(key)
	require.True(t, ok)
	require.NotEqual(t, nil, val)
	require.Equal(t, 1, val.(int))

	time.Sleep(TTL)

	_, ok = c.Get(key)
	assert.False(t, ok)
}

func Test_CacheOverwriteKey(t *testing.T) {
	c := makeCache(t, TTL)

	assert.True(t, c.Set(key, 1))
	time.Sleep(wait)
	assert.True(t, c.Set(key, 2))
	time.Sleep(wait)

	val, ok := c.Get(key)
	require.True(t, ok)
	require.NotEqual(t, nil, val)
	assert.Equal(t, 2, val.(int))
}

func Test_CacheSetGetVerifyCounters(t *testing.T) {
	c := makeCache(t, TTL)
	ok := c.Set(key, "data")
	require.True(t, ok)

	time.Sleep(wait)

	value, ok := c.Get(key)
	assert.True(t, ok)
	assert.Equal(t, "data", value.(string))
	assert.Equal(t, 1, int(c.Sets()))
	assert.Equal(t, 1, int(c.Hits()))
	assert.Equal(t, 0, int(c.Misses()))
	assert.Equal(t, 1, int(c.Gets()))
	assert.Equal(t, 1, int(c.perFuncMetrics[cacheFuncKey].gets))
	assert.Equal(t, 1, int(c.perFuncMetrics[cacheFuncKey].hits))
}

func Test_CacheMissVerifyCounters(t *testing.T) {
	c := makeCache(t, TTL)

	value, ok := c.Get(key)
	require.False(t, ok)
	require.Nil(t, value)
	assert.Equal(t, 0, int(c.Sets()))
	assert.Equal(t, 0, int(c.Hits()))
	assert.Equal(t, 1, int(c.Misses()))
	assert.Equal(t, 1, int(c.Gets()))
	assert.Equal(t, 1, int(c.perFuncMetrics[cacheFuncKey].gets))
	assert.Equal(t, 0, int(c.perFuncMetrics[cacheFuncKey].hits))
}

// nolint: gomnd
func Test_CacheTestPerFuncKeyCounters(t *testing.T) {
	c := makeCache(t, TTL)

	assert.True(t, c.Set(key, 1))
	assert.True(t, c.Set(key2, 2))
	time.Sleep(wait)

	loops := 10
	expectedGets := loops * 2
	expectedHits := loops * 2

	for i := 0; i < 10; i++ {
		val, ok := c.Get(key)
		require.True(t, ok)
		require.Equal(t, 1, val.(int))
		val, ok = c.Get(key2)
		require.True(t, ok)
		require.Equal(t, 2, val.(int))
	}

	assert.Equal(t, expectedGets, int(c.Gets()))
	assert.Equal(t, expectedHits, int(c.Hits()))
	assert.Equal(t, expectedGets/2, int(c.perFuncMetrics[key.FuncName()].gets))
	assert.Equal(t, expectedHits/2, int(c.perFuncMetrics[key.FuncName()].hits))
	assert.Equal(t, expectedGets/2, int(c.perFuncMetrics[key2.FuncName()].gets))
	assert.Equal(t, expectedHits/2, int(c.perFuncMetrics[key2.FuncName()].hits))
}
