package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
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
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	// let's make some random bytes
	testdata := []byte("This is a test content!")
	testkey := "testkey"

	err := s.WriteStream(testkey, bytes.NewReader(testdata))
	assert.Nil(t, err)

	reader, err := s.ReadStream(testkey)
	assert.Nil(t, err)

	b, err := io.ReadAll(reader)
	assert.Nil(t, err)
	assert.Equal(t, testdata, b)

	s.Delete(testkey)
}

func TestStoreDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	// let's make some random bytes
	testkey := "testkey"
	testdata := []byte("This is a test content!")

	err := s.WriteStream(testkey, bytes.NewReader(testdata))

	assert.Nil(t, err)

	err = s.Delete(testkey)
	assert.Nil(t, err)

	_, err = s.ReadStream(testkey)

	log.Printf("%+v", err)

	assert.NotNil(t, err)
}

func TestStoreHas(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	testkey := "testkey"
	testdata := []byte("This is a test content!")

	err := s.WriteStream(testkey, bytes.NewReader(testdata))
	assert.Nil(t, err)

	ok := s.Has(testkey)
	assert.True(t, ok)

	err = s.Delete(testkey)
	assert.Nil(t, err)

	ok = s.Has(testkey)
	assert.False(t, ok)
}
