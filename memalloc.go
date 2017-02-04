// Package memalloc provides a generic memory allocator.
package memalloc

import "unsafe"

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

// MemAllocator wraps an Allocator and uses it to create
// unsafe.Pointers.
type MemAllocator struct {
	Start     unsafe.Pointer
	Size      int
	Allocator Allocator
}

// Contains checks if a pointer is within the allocator's
// controlled memory region.
func (m *MemAllocator) Contains(p unsafe.Pointer) bool {
	return uintptr(p) < uintptr(m.Start)+uintptr(m.Size) && uintptr(p) >= uintptr(m.Start)
}

// Alloc allocates a pointer.
func (m *MemAllocator) Alloc(size int) (unsafe.Pointer, error) {
	addr, err := m.Allocator.Alloc(size)
	if err != nil {
		return nil, err
	}
	return unsafe.Pointer(uintptr(m.Start) + uintptr(addr)), nil
}

// Free frees a pointer.
func (m *MemAllocator) Free(ptr unsafe.Pointer) {
	if !m.Contains(ptr) {
		panic("pointer out of bounds")
	}
	idx := int(uintptr(ptr) - uintptr(m.Start))
	m.Allocator.Free(idx)
}
