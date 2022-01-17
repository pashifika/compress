// Package compress
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
package compress

import (
	"errors"
	"io"
	"io/fs"
	"os"

	"golang.org/x/text/encoding"
)

const DefaultArchiverRoot = "."

type ArchiverFile interface {
	fs.FileInfo
	fs.ReadDirFile
	fs.DirEntry
	io.Writer

	Root() string
}

type Decoder interface {
	// Name is get Decoder name.
	Name() string

	// SetRootInfo is set archiver file info to Decoder.
	SetRootInfo(info os.FileInfo)

	// SetCharset option support(only work in zip).
	SetCharset(charset []encoding.Encoding, skipErr bool)

	// OpenReader open the archive specified by name and returns io.ReadCloser.
	//
	// * 7z / rar is support part files.
	OpenReader(path string) (fs.FS, error)

	// OpenReaderWithPassword will open the archive file specified by name using password as
	// the basis of the decryption key and returns io.ReadCloser.
	//
	// * 7z / rar is support part files.
	OpenReaderWithPassword(path, pwd string) (fs.FS, error)

	// GetDirEntries get the archive path entries (if you know).
	GetDirEntries(path string, n int) ([]fs.DirEntry, error)

	// Close closes the archive file, rendering it unusable for I/O.
	Close() error

	// Reset is reset Decoder memory cache.
	Reset()
}

type Encoder interface {
	// Name is get Encoder name.
	Name() string

	SetCompressedExt(ext map[string]struct{})

	// Create use encoder name to create new archive file
	Create(w io.Writer, entries []ArchiverFile) error

	// Close closes the archive file, writing it unusable for I/O.
	Close() error

	// Reset is reset Encoder memory cache.
	Reset()
}

// errors

var (
	ErrUnknownEncoder   = errors.New("unknown encoder")
	ErrUnknownArchiver  = errors.New("unknown archiver file")
	ErrWriterNotSupport = errors.New("writer is not supported")

	// ErrDirIndexTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
	ErrDirIndexTooLarge = errors.New("DirIndex.slice: too large")
)
