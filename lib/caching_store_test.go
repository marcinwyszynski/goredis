package lib

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type cachingStoreTestSuite struct {
	suite.Suite

	authority *mockStore
	cache     *mockStore

	sut *CachingStore
}

func (c *cachingStoreTestSuite) SetupTest() {
	c.authority = new(mockStore)
	c.cache = new(mockStore)
	c.sut = NewCachingStore(c.authority, c.cache)
}

func (c *cachingStoreTestSuite) TestGet_KnownMissing() {
	const key = "key"

	c.sut.KnownMissing[key] = struct{}{}

	ret, found, err := c.sut.Get(key)

	c.Empty(ret)
	c.False(found)
	c.NoError(err)
}

func (c *cachingStoreTestSuite) TestGet_FoundInCache() {
	const key = "key"
	const value = "value"

	c.cache.On("Get", key).Return(value, true, nil)

	ret, found, err := c.sut.Get(key)

	c.Equal(value, ret)
	c.True(found)
	c.NoError(err)

	c.Empty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestGet_ErrorQueryingCache() {
	const key = "key"

	c.cache.On("Get", key).Return("", false, errors.New("bacon"))

	ret, found, err := c.sut.Get(key)

	c.Empty(ret)
	c.False(found)
	c.EqualError(err, "could not retrieve value from cache: bacon")

	c.Empty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestGet_ErrorQueryingAuthority() {
	const key = "key"

	c.cache.On("Get", key).Return("", false, nil)
	c.authority.On("Get", key).Return("", false, errors.New("bacon"))

	ret, found, err := c.sut.Get(key)

	c.Empty(ret)
	c.False(found)
	c.EqualError(err, "could not retrieve value from authority: bacon")

	c.Empty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestGet_NotFoundInAuthority() {
	const key = "key"

	c.cache.On("Get", key).Return("", false, nil)
	c.authority.On("Get", key).Return("", false, nil)

	ret, found, err := c.sut.Get(key)

	c.Empty(ret)
	c.False(found)
	c.NoError(err)

	c.NotEmpty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestGet_FoundInAuthority() {
	const key = "key"
	const value = "value"

	c.cache.
		On("Get", key).Return("", false, nil).
		On("Set", key, value).Return(nil)

	c.authority.On("Get", key).Return(value, true, nil)

	ret, found, err := c.sut.Get(key)

	c.Equal(value, ret)
	c.True(found)
	c.NoError(err)

	c.Empty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestGet_ErrorSettingCache() {
	const key = "key"
	const value = "value"

	c.cache.
		On("Get", key).Return("", false, nil).
		On("Set", key, value).Return(errors.New("bacon"))

	c.authority.On("Get", key).Return(value, true, nil)

	ret, found, err := c.sut.Get(key)

	c.Equal(value, ret)
	c.True(found)
	c.EqualError(err, "could not set value in cache: bacon")
}

func (c *cachingStoreTestSuite) TestSet_OK() {
	const key = "key"
	const value = "value"

	c.sut.KnownMissing[key] = struct{}{}

	c.authority.On("Set", key, value).Return(nil)
	c.cache.On("Set", key, value).Return(nil)

	c.NoError(c.sut.Set(key, value))
	c.Empty(c.sut.KnownMissing)
}

func (c *cachingStoreTestSuite) TestSet_AuthorityError() {
	const key = "key"
	const value = "value"

	c.authority.On("Set", key, value).Return(errors.New("bacon"))

	c.EqualError(c.sut.Set(key, value), "could not set value in authority: bacon")
}

func (c *cachingStoreTestSuite) TestSet_CacheError() {
	const key = "key"
	const value = "value"

	c.authority.On("Set", key, value).Return(nil)
	c.cache.On("Set", key, value).Return(errors.New("bacon"))

	c.EqualError(c.sut.Set(key, value), "could not set value in cache: bacon")
}

func TestCachingStore(t *testing.T) {
	suite.Run(t, new(cachingStoreTestSuite))
}
