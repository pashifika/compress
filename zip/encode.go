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
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/pashifika/compress"
	"github.com/pashifika/compress/internal/std_zip"
)

type WriteCloser struct {
	extensions map[string]struct{}
	close      func() error
}

func (wc *WriteCloser) Name() string { return _zipName }

func (wc *WriteCloser) SetCompressedExt(ext map[string]struct{}) { wc.extensions = ext }

func (wc *WriteCloser) Create(w io.Writer, entries []compress.ArchiverFile) error {
	zip := std_zip.NewWriter(w)
	//goland:noinspection ALL
	defer zip.Close()

	for i, entry := range entries {
		header, err := std_zip.FileInfoHeader(entry)
		if err != nil {
			return err
		}
		root := entry.Root()
		if entry.IsDir() {
			if !strings.HasSuffix(root, "/") {
				header.Name += "/" // required
			}
			header.Method = std_zip.Store
		} else {
			ext := strings.ToLower(path.Ext(root))
			if _, ok := wc.extensions[ext]; ok {
				header.Method = std_zip.Store
			} else {
				header.Method = std_zip.Deflate
			}
			header.Name = root
		}
		zw, err := zip.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("zip creating header for file [%d] %s\n  error: %w", i, header.Name, err)
		}

		// directories have no file body
		if entry.IsDir() {
			continue
		}
		_, err = io.Copy(zw, entry)
		if err != nil {
			return fmt.Errorf("writing file [%d] %s\n  error: %w", i, header.Name, err)
		}
		//if header.Method == std_zip.Store && n != entry.Size() {
		//	return fmt.Errorf("writing file [%d] %s\n  size error: (%d/%d)", i, header.Name, n, entry.Size())
		//}
	}

	return zip.Flush()
}

func (wc *WriteCloser) Close() error {
	if wc.close != nil {
		return wc.close()
	}
	return nil
}

func (wc *WriteCloser) Reset() {
	*wc = WriteCloser{}
}

func init() {
	compress.RegisterEncoder(&WriteCloser{})
}
