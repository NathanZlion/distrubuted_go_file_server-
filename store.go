package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "p2pnetworkStore"

type PathTransformFunc func(string) Pathkey

// A function to transform some key to a unique
// path separated with /
func CASPathTransformFunc(key string) Pathkey {
	hashBytes := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hashBytes[:])

	blockSize := 5

	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := range sliceLen {
		paths[i] = hashStr[i*blockSize : (i+1)*blockSize]
	}

	return Pathkey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

func DefaultPathTransformFunc(key string) Pathkey {
	return Pathkey{
		PathName: key,
		FileName: key,
	}
}

type Pathkey struct {
	PathName string
	FileName string
}

type StoreOpts struct {
	// Root path name for the folder that contains all the files
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}

	return &Store{StoreOpts: opts}
}

func (pathKey *Pathkey) FullPath(root string) string {
	return root + "/" + pathKey.PathName + "/" + pathKey.FileName
}

func (s *Store) WriteStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	if err := os.MkdirAll(s.Root+"/"+pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	pathWithFileName := pathKey.FullPath(s.Root)

	fileWriter, err := os.Create(pathWithFileName)
	defer fileWriter.Close()

	if err != nil {
		return err
	}

	// opening the file
	written, err := io.Copy(fileWriter, r)

	if err != nil {
		fmt.Printf("Error while copying to file writer %v \n", err)
		return err
	}

	fmt.Printf("Written (%d) bytes to disk at path %s \n", written, pathWithFileName)

	return nil
}

func (s *Store) ReadStream(key string) (io.Reader, error) {
	pathKey := s.PathTransformFunc(key)

	fileReader, err := os.Open(pathKey.FullPath(s.Root))
	defer fileReader.Close()

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, fileReader)

	return buf, err
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleted [%s] from disk", pathKey.FileName)
	}()

	return os.RemoveAll(pathKey.FullPath(s.Root))
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)

	fullPath := pathKey.FullPath(s.Root)

	_, err := os.Stat(fullPath)

	return !os.IsNotExist(err)
}
