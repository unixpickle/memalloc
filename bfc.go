package memalloc

import (
	"errors"
)

// bfc is a best-fit with coalescing allocator.
// During allocations, the smallest fitting free chunk is
// selected.
// During frees, neighboring free chunks are re-joined to
// prevent memory fragmentation.
type bfc struct {
	size  int
	align int
	free  []*freeChunk
	sizes map[int]int
}

// NewBFC creates an Allocator that uses a best-fit with
// coalescing algorithm.
//
// The resulting allocator will always produce buffers
// aligned to the given alignment.
func NewBFC(size, align int) Allocator {
	return &bfc{
		align: align,
		size:  downAlign(size, align),
		free:  []*freeChunk{{0, size}},
		sizes: map[int]int{},
	}
}

func (b *bfc) Alloc(size int) (addr int, err error) {
	if size == 0 {
		size = 1
	}
	size = upAlign(size, b.align)
	var smallestFit *freeChunk
	var fitIdx int
	for i, f := range b.free {
		if f.size >= size {
			if smallestFit == nil || smallestFit.size > f.size {
				fitIdx = i
				smallestFit = f
			}
		}
	}
	if smallestFit == nil {
		return 0, errors.New("alloc: out of memory")
	}
	res := smallestFit.start
	remaining := smallestFit.size - size
	if remaining > 0 {
		smallestFit.size = remaining
		smallestFit.start += size
	} else {
		b.free[fitIdx] = b.free[len(b.free)-1]
		b.free = b.free[:len(b.free)-1]
	}
	b.sizes[res] = size
	return res, nil
}

func (b *bfc) Free(addr int) {
	size := b.sizes[addr]
	delete(b.sizes, addr)

	var lastFree, nextFree *freeChunk
	var nextFreeIdx int
	for i, f := range b.free {
		if f.start == addr+size {
			nextFree = f
			nextFreeIdx = i
		} else if f.start+f.size == addr {
			lastFree = f
		}
	}

	if lastFree == nil && nextFree == nil {
		b.free = append(b.free, &freeChunk{start: addr, size: size})
	} else if lastFree == nil {
		nextFree.start = addr
		nextFree.size += size
	} else {
		lastFree.size += size
		if nextFree != nil {
			lastFree.size += nextFree.size
			b.free[nextFreeIdx] = b.free[len(b.free)-1]
			b.free = b.free[:len(b.free)-1]
		}
	}
}

type freeChunk struct {
	start int
	size  int
}

func downAlign(size, align int) int {
	size -= (size % align)
	return size
}

func upAlign(size, align int) int {
	if size%align == 0 {
		return size
	} else {
		return downAlign(size, align) + align
	}
}
