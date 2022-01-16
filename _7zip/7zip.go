// Package _7zip
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
package _7zip

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	"golang.org/x/text/encoding"

	"github.com/pashifika/compress"
)

type ReadCloser struct {
	_7z     *sevenzip.ReadCloser
	entries map[string]*compress.DirIndex
	dirs    map[string]int
	files   map[string]int
	index   []*File

	root fs.FileInfo
}

func (rc *ReadCloser) Name() string { return "7zip" }

func (rc *ReadCloser) SetRootInfo(info os.FileInfo) { rc.root = info }

func (rc *ReadCloser) SetCharset(_ []encoding.Encoding, _ bool) {}

// OpenReader will open the 7-zip file specified by name and return a
// ReadCloser. If name has a ".001" suffix it is assumed there are multiple
// volumes and each sequential volume will be opened.
func (rc *ReadCloser) OpenReader(path string) (fs.FS, error) {
	return rc.OpenReaderWithPassword(path, "")
}

// OpenReaderWithPassword will open the 7-zip file specified by name using
// password as the basis of the decryption key and return a ReadCloser. If
// name has a ".001" suffix it is assumed there are multiple volumes and each
// sequential volume will be opened.
func (rc *ReadCloser) OpenReaderWithPassword(path, pwd string) (fs.FS, error) {
	_7zip, err := sevenzip.OpenReaderWithPassword(path, pwd)
	if err != nil {
		return nil, err
	}

	maxIdx := len(_7zip.File) + 1
	res := &ReadCloser{_7z: _7zip,
		entries: map[string]*compress.DirIndex{
			compress.DefaultArchiverRoot: compress.NewDirEntries(),
		},
		dirs:  map[string]int{},
		files: map[string]int{},
		index: make([]*File, maxIdx),
		root:  rc.root,
	}
	for idx, file := range res._7z.File {
		mode := file.FileHeader.Mode()
		entry := &File{f: file, size: 0, mode: mode}
		if mode.IsDir() {
			entry.isDir = true
			entry.name = strings.TrimRight(file.Name, "/")
			entry.dirEntries = res.GetDirEntries
			res.dirs[entry.name] = idx
			res.entries[compress.DefaultArchiverRoot].Add(idx)
		} else {
			entry.name = file.Name
			entry.size = int64(file.FileHeader.UncompressedSize)
			res.files[entry.name] = idx
			// Add index to dir entries
			dir := filepath.Dir(entry.name)
			if _, ok := res.entries[dir]; !ok {
				res.entries[dir] = compress.NewDirEntries()
			}
			res.entries[dir].Add(idx)
		}
		res.index[idx] = entry
	}
	// Set root info
	rootIdx := maxIdx - 1
	res.dirs[compress.DefaultArchiverRoot] = rootIdx
	res.index[rootIdx] = &File{
		name:       compress.DefaultArchiverRoot,
		mode:       res.root.Mode() + os.ModeDir,
		isDir:      true,
		dirEntries: res.GetDirEntries,
	}

	return res, nil
}

// Open opens the named file in the 7-zip file, using the semantics of fs.FS.Open:
// paths are always slash separated, with no leading / or ../ elements.
func (rc *ReadCloser) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	idx, ok := rc.dirs[name]
	if ok {
		return rc.getFile(idx)
	}
	idx, ok = rc.files[name]
	if ok {
		return rc.getFile(idx)
	}
	return nil, &fs.PathError{Op: "info", Path: name, Err: fs.ErrNotExist}
}

func (rc *ReadCloser) GetDirEntries(path string, n int) ([]fs.DirEntry, error) {
	var (
		entries []fs.DirEntry
		dir     *File
	)

	if dIdx, ok := rc.dirs[path]; !ok {
		return nil, fs.ErrNotExist
	} else {
		if n > 0 {
			var err error
			dir, err = rc.getFile(dIdx)
			if err != nil {
				return nil, err
			}
			n = dir.dirReadAt + n
		}
	}
	if di, ok := rc.entries[path]; ok {
		entries = make([]fs.DirEntry, di.Len())
		for idx, fIdx := range di.Entries() {
			file, err := rc.getFile(fIdx)
			if err != nil {
				return nil, err
			}
			entries[idx] = file
			if n > 0 && idx >= n {
				break
			}
		}
	} else {
		return nil, fs.ErrNotExist
	}

	if len(entries) == 0 {
		return nil, io.EOF
	}
	if dir != nil {
		dir.dirReadAt = n
	}
	return entries, nil
}

func (rc *ReadCloser) getFile(idx int) (*File, error) {
	if idx > len(rc.index) || idx < 0 {
		return nil, fs.ErrInvalid
	}
	file := rc.index[idx]
	if !file.isDir {
		err := file.OpenFile()
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}

func (rc *ReadCloser) Reset() {
	*rc = ReadCloser{}
}

// Close closes the 7-zip file or volumes, rendering them unusable for I/O.
func (rc *ReadCloser) Close() error {
	if rc._7z != nil {
		err := rc._7z.Close()
		return err
	}
	if rc != nil {
		rc.Reset()
	}
	return nil
}
