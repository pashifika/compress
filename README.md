compress: a go compress library for fs.FS interface
===================================================


| Format | Test  | Charset | Decoder | Encoder | Password | Info                                                                                           |
|--------|-------|---------|---------|---------|----------|------------------------------------------------------------------------------------------------|
| zip    | local | true    | true    | true    | false    | used go std                                                                                    |
| rar    | local | false   | true    | false   | true     | [rardecode/v2](http://github.com/nwaples/rardecode)                                            |
| 7zip   | false | false   | true    | false   | true     | not work in big file(>10M)<br/>github.com/ulikunitz/xz/lzma.(*rangeDecoder).DecodeBit too slow |




## use

```
go get github.com/pashifika/compress
```

Example:
--------
```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"

	"github.com/pashifika/compress"
	_ "github.com/pashifika/compress/rar"
	_ "github.com/pashifika/compress/zip"
)

func main() {
	// set charset to decode zip header name
	fsys := &compress.FileSystem{
		Charset:     []encoding.Encoding{japanese.ShiftJIS},
		SkipCharErr: false,
	}
	path := "you test archive file path"
	rc, err := fsys.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fsys.Close()

	err = fs.WalkDir(rc, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch path {
			case ".git":
				return fs.SkipDir
			case compress.DefaultArchiverRoot:
				return nil
			default:
				fmt.Println("dir:", path)
				return nil
			}
		}
		if !d.IsDir() {
			af, err := rc.Open(path)
			if err != nil {
				panic(err)
			}
			buf := bytes.Buffer{}
			n, err := io.Copy(&buf, af)
			if err != nil {
				panic(err)
			}
			fmt.Println("file:", path)
			fmt.Println(n, err, n, buf.Len())
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```
