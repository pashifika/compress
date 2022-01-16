// Package zip
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
package zip

import (
	"io/fs"
	"os"

	"github.com/pashifika/compress/internal/std_zip"

	"github.com/pashifika/compress"
)

type ReadCloser struct {
	zip *std_zip.ReadCloser
}

func (rc *ReadCloser) Name() string { return "zip" }

func (rc *ReadCloser) SetRootInfo(_ os.FileInfo) {}

func (rc *ReadCloser) OpenReader(path string) (fs.FS, error) {
	return rc.OpenReaderWithPassword(path, "")
}

func (rc *ReadCloser) OpenReaderWithPassword(path, _ string) (fs.FS, error) {
	z, err := std_zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	rc.zip = z
	return z, nil
}

func (rc *ReadCloser) GetDirEntries(_ string, _ int) ([]fs.DirEntry, error) {
	return nil, nil
}

func (rc *ReadCloser) Reset() {
	*rc = ReadCloser{}
}

func (rc *ReadCloser) Close() error {
	if rc.zip != nil {
		return rc.zip.Close()
	}
	return nil
}

func init() {
	compress.RegisterDecoder(&ReadCloser{})
}
