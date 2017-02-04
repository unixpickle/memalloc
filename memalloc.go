// Package memalloc provides a generic memory allocator.
package memalloc

// An Allocator allocates ranges of integers.
//
// An Allocator controls an implicit buffer that starts at
// address 0 and has a pre-determined size.
// Allocation reserves sub-ranges in this buffer.
// Freeing allows sub-ranges to be allocated again.
type Allocator interface {
	Alloc(size int) (addr int, err error)
	Free(addr int)
}
