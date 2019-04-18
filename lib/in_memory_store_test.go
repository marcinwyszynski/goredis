package lib

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type inMemoryStoreTestSuite struct {
	suite.Suite

	sut Store
}

func (i *inMemoryStoreTestSuite) SetupTest() {
	i.sut = NewInMemoryStore()
}

func (i *inMemoryStoreTestSuite) TestSuccessfulRoundTrip() {
	const key = "key"
	const val = "val"

	i.NoError(i.sut.Set(key, val))

	ret, found, err := i.sut.Get(key)
	i.Equal(val, ret)
	i.True(found)
	i.NoError(err)
}

func (i *inMemoryStoreTestSuite) TestFailedGet() {
	const key = "key"

	ret, found, err := i.sut.Get(key)
	i.Empty(ret)
	i.False(found)
	i.NoError(err)
}

func TestInMemoryStore(t *testing.T) {
	suite.Run(t, new(inMemoryStoreTestSuite))
}
