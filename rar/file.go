// Package rar
/*
 * Version: 1.0.0
 * Copyright (c) 2022. Pashifika
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package rar

import (
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/nwaples/rardecode/v2"

	"github.com/pashifika/compress"
)

type File struct {
	header *rardecode.FileHeader
	name   string
	isDir  bool
	size   int64
	mode   fs.FileMode

	dirReadAt  int
	dirEntries func(path string, n int) ([]fs.DirEntry, error)
	fileOpen   func() (io.ReadCloser, error)
	rcRead     func(p []byte) (n int, err error)
	close      func() error
}

func (f *File) Root() string { return f.header.Name }

func (f *File) IsDir() bool { return f.isDir }

func (f *File) Size() int64 { return f.size }

func (f *File) Write(_ []byte) (n int, err error) {
	return 0, compress.ErrWriterNotSupport
}

func (f *File) OpenFile() error {
	rc, err := f.fileOpen()
	if err != nil {
		return err
	}
	f.rcRead = rc.Read
	f.close = func() error {
		err := rc.Close()
		f.rcRead = nil
		f.close = nil
		f.fileOpen = nil
		f.header = nil
		return err
	}
	return nil
}

// ------ to fs.FileInfo ------

func (f *File) Mode() fs.FileMode {
	if f.header == nil {
		return f.mode
	}
	return f.header.Mode()
}

func (f *File) ModTime() time.Time {
	if f.header == nil {
		return time.Time{}
	}
	return f.header.ModificationTime.UTC()
}

func (f *File) Sys() interface{} {
	if f.header == nil {
		return nil
	}
	return f.header
}

// ------ to fs.File ------

func (f *File) Stat() (fs.FileInfo, error) { return f, nil }

func (f *File) Read(b []byte) (int, error) { return f.rcRead(b) }

func (f *File) Close() error {
	if f.close != nil {
		return f.close()
	}
	return nil
}

// ------ to fs.DirEntry ------

func (f *File) Name() string {
	if f.isDir {
		return f.name
	}
	return path.Base(f.name)
}

func (f *File) Type() fs.FileMode { return f.mode }

func (f *File) Info() (fs.FileInfo, error) { return f, nil }

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.dirEntries != nil {
		return f.dirEntries(f.name, n)
	}
	return nil, fs.ErrNotExist
}
