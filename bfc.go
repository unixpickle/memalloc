package memalloc

import (
	"errors"

	"github.com/unixpickle/splaytree"
)

// bfc is a best-fit with coalescing allocator.
// During allocations, the smallest fitting free chunk is
// selected.
// During frees, neighboring free chunks are re-joined to
// prevent memory fragmentation.
type bfc struct {
	size          int
	align         int
	free          *splaytree.Tree[*freeNode]
	usedSizes     map[int]int
	freeToSize    map[int]int
	freeEndToSize map[int]int
}

// NewBFC creates an Allocator that uses a best-fit with
// coalescing algorithm.
//
// The resulting allocator will always produce buffers
// aligned to the given alignment.
func NewBFC(size, align int) Allocator {
	b := &bfc{
		align: align,
		size:  downAlign(size, align),
		free:  &splaytree.Tree[*freeNode]{},

		usedSizes:     map[int]int{},
		freeToSize:    map[int]int{},
		freeEndToSize: map[int]int{},
	}
	b.freeToSize[0] = b.size
	b.freeEndToSize[b.size] = b.size
	b.free.Insert(&freeNode{start: 0, size: b.size})
	return b
}

func (b *bfc) Alloc(size int) (addr int, err error) {
	if size == 0 {
		size = 1
	}
	size = upAlign(size, b.align)
	fit := b.smallestFit(b.free.Root, size)
	if fit == nil {
		return 0, errors.New("alloc: out of memory")
	}
	b.free.Delete(fit)
	delete(b.freeToSize, fit.start)
	delete(b.freeEndToSize, fit.start+fit.size)

	res := fit.start
	remaining := fit.size - size
	if remaining > 0 {
		fit.size = remaining
		fit.start += size
		b.free.Insert(fit)
		b.freeToSize[fit.start] = remaining
		b.freeEndToSize[fit.start+remaining] = remaining
	}
	b.usedSizes[res] = size
	return res, nil
}

func (b *bfc) Free(addr int) {
	size := b.usedSizes[addr]
	delete(b.usedSizes, addr)

	var lastFree, nextFree *freeNode
	if lastSize, ok := b.freeEndToSize[addr]; ok {
		lastFree = &freeNode{start: addr - lastSize, size: lastSize}
		b.free.Delete(lastFree)
		delete(b.freeEndToSize, addr)
		delete(b.freeToSize, lastFree.start)
	}
	if nextSize, ok := b.freeToSize[addr+size]; ok {
		nextFree = &freeNode{start: addr + size, size: nextSize}
		b.free.Delete(nextFree)
		delete(b.freeEndToSize, nextFree.start+nextFree.size)
		delete(b.freeToSize, nextFree.start)
	}

	if lastFree == nil {
		lastFree = &freeNode{start: addr, size: size}
	} else {
		lastFree.size += size
	}
	if nextFree != nil {
		lastFree.size += nextFree.size
	}

	b.freeToSize[lastFree.start] = lastFree.size
	b.freeEndToSize[lastFree.start+lastFree.size] = lastFree.size
	b.free.Insert(lastFree)
}

func (b *bfc) smallestFit(n *splaytree.Node[*freeNode], size int) *freeNode {
	if n == nil {
		return nil
	}
	val := n.Value
	if val.size == size {
		return val
	}
	if val.size > size {
		if res := b.smallestFit(n.Left, size); res != nil {
			return res
		}
		return val
	} else {
		return b.smallestFit(n.Right, size)
	}
}

type freeNode struct {
	start int
	size  int
}

func (f *freeNode) Compare(v2 *freeNode) int {
	f2 := v2
	if f.size < f2.size {
		return -1
	} else if f.size > f2.size {
		return 1
	} else if f.start < f2.start {
		return -1
	} else if f.start > f2.start {
		return 1
	} else {
		return 0
	}
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
