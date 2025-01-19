package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "some key"
	hashPaths := CASPathTransformFunc(key)

	// assert same result is produced when called with the same key
	assert.Equal(t, CASPathTransformFunc(key), hashPaths)

	// expected value
	expectedHashPath := "ab0d8/e0ce5/8e6fa/9d1b2/30d25/f2ea0/b44a5/1ebd4"
	expectedHashMismatchMsg := fmt.Sprintf("Expected hashstring %s but got %s \n", expectedHashPath, hashPaths)
	assert.Equal(t, expectedHashPath, hashPaths.PathName, expectedHashMismatchMsg)

	expectedHashPath = "ab0d8e0ce58e6fa9d1b230d25f2ea0b44a51ebd4"
	expectedHashMismatchMsg = fmt.Sprintf("Expected hashstring %s but got %s \n", expectedHashPath, hashPaths)
	assert.Equal(t, expectedHashPath, hashPaths.FileName, expectedHashMismatchMsg)
}

func TestStore(t *testing.T) {
	s := newStoreHelper()
	defer tearDown(t, s)

	// let's make some random bytes
	testkey := "testkey"
	testdata := []byte("This is a test content!")

	_, err := s.WriteStream(testkey, bytes.NewReader(testdata))
	assert.Nil(t, err)

	reader, err := s.ReadStream(testkey)
	assert.Nil(t, err)

	b, err := io.ReadAll(reader)
	assert.Nil(t, err)
	assert.Equal(t, testdata, b)

	s.Delete(testkey)
}

func TestStoreIntegration(t *testing.T) {
	s := newStoreHelper()
	defer tearDown(t, s)
	testdata := []byte("This is a test content!")

	for round := range 5 {
		testkey := fmt.Sprintf("foo%v", round)

		// Write
		_, err := s.WriteStream(testkey, bytes.NewReader(testdata))
		assert.Nil(t, err)

		// Read
		reader, err := s.ReadStream(testkey)
		assert.Nil(t, err)
		b, err := io.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, testdata, b)

		// Has
		ok := s.Has(testkey)
		assert.True(t, ok)

		// Delete
		err = s.Delete(testkey)
		assert.Nil(t, err)

		ok = s.Has(testkey)
		assert.False(t, ok)
	}
}

func newStoreHelper() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	return NewStore(opts)
}

// we want to clear everything in the store
func tearDown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
