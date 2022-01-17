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

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64
const maxInt = int(^uint(0) >> 1)

type DirIndex struct {
	slice []int
	off   int64
}

func NewDirEntries() *DirIndex {
	return &DirIndex{slice: []int{}}
}

func (d *DirIndex) Add(idx int) {
	d.slice = append(d.slice, idx)
}

func (d *DirIndex) Entries() []int { return d.slice }

// Len returns the number of slice of the unread portion of the DirIndex;
func (d *DirIndex) Len() int { return len(d.slice) - int(d.off) }

// Reset resets the DirIndex to be empty
func (d *DirIndex) Reset() {
	d.slice = d.slice[:0]
	d.off = 0
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal DirIndex only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (d *DirIndex) tryGrowByReslice(n int) (int, bool) {
	if l := len(d.slice); n <= cap(d.slice)-l {
		d.slice = d.slice[:l+n]
		return l, true
	}
	return 0, false
}

// grow the buffer to guarantee space for n more slice.
// It returns the index where bytes should be written.
// If the DirIndex can't grow it will panic with ErrDirIndexTooLarge.
func (d *DirIndex) grow(n int) int {
	m := d.Len()
	// If DirIndex is empty, reset to recover space.
	if m == 0 && d.off != 0 {
		d.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := d.tryGrowByReslice(n); ok {
		return i
	}
	if d.slice == nil && n <= smallBufferSize {
		d.slice = make([]int, n, smallBufferSize)
		return 0
	}
	c := cap(d.slice)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(d.slice, d.slice[d.off:])
	} else if c > maxInt-c-n {
		panic(ErrDirIndexTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		index := makeEntries(2*c + n)
		copy(index, d.slice[d.off:])
		d.slice = index
	}
	// Restore d.off and len(d.slice).
	d.off = 0
	d.slice = d.slice[:m+n]
	return m
}

// makeEntries allocates a slice of size n. If the allocation fails, it panics
// with ErrDirIndexTooLarge.
func makeEntries(n int) []int {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrDirIndexTooLarge)
		}
	}()
	return make([]int, n)
}
