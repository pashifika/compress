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
	"io"
	"io/fs"
	"os"

	"golang.org/x/text/encoding"
)

type FileSystem struct {
	Charset     []encoding.Encoding
	SkipCharErr bool

	close func() error
}

// Open cannot work with OpenWithPwd at the same time
func (fs *FileSystem) Open(path string) (fs.FS, error) { return fs.OpenWithPwd(path, "") }

// OpenWithPwd cannot work with Open at the same time
func (fs *FileSystem) OpenWithPwd(path, pwd string) (fs.FS, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return os.DirFS(path), nil
	}

	// Decoder supported archiver file
	for _, decoder := range decoders {
		decoder.SetRootInfo(info)
		if fs.Charset != nil {
			decoder.SetCharset(fs.Charset, fs.SkipCharErr)
		}
		rc, err := decoder.OpenReaderWithPassword(path, pwd)
		if err != nil {
			continue
		}
		fs.close = decoder.Close
		return rc, nil
	}
	return nil, ErrUnknownArchiver
}

// CreateArchiverFile save archiver entries to disk
func (fs *FileSystem) CreateArchiverFile(encode string, w io.Writer, entries []ArchiverFile) error {
	encoder, ok := encoders[encode]
	if !ok {
		return ErrUnknownEncoder
	}
	return encoder.Create(w, entries)
}

func (fs *FileSystem) Close() error {
	if fs.close != nil {
		return fs.close()
	}
	return nil
}
