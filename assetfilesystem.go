package main

// http://qiita.com/sfujiwara/items/b84cc2a7b326b04e0edb

import (
	"bytes"
	"net/http"
	"os"
	"strings"
)

type AssetFileSystem struct{}

type AssetFile struct {
	*bytes.Reader
	os.FileInfo
}

func (fs AssetFileSystem) Open(name string) (http.File, error) {
	path := strings.TrimLeft(name, "/")
	data, err := Asset(path)
	if err != nil {
		return nil, err
	}
	info, err := AssetInfo(path)
	if err != nil {
		return nil, err
	}
	file := &AssetFile{
		bytes.NewReader(data),
		info,
	}
	return file, nil
}

func (f *AssetFile) Close() error {
	return nil
}

func (f *AssetFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (f *AssetFile) Stat() (os.FileInfo, error) {
	return f.FileInfo, nil
}
