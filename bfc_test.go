package memalloc

import (
	"math/rand"
	"testing"
)

func TestBFCManual(t *testing.T) {
	allocator := NewBFC(512, 16)
	addr := mustAlloc(allocator.Alloc(5))
	if addr != 0 {
		t.Errorf("expected 5 but got %d", addr)
	}
	addr = mustAlloc(allocator.Alloc(17))
	if addr != 16 {
		t.Errorf("expected 16 but got %d", addr)
	}
	addr = mustAlloc(allocator.Alloc(0))
	if addr != 48 {
		t.Errorf("expected 32 but got %d", addr)
	}
	addr = mustAlloc(allocator.Alloc(15))
	if addr != 64 {
		t.Errorf("expected 64 but got %d", addr)
	}
	allocator.Free(48)
	addr = mustAlloc(allocator.Alloc(15))
	if addr != 48 {
		t.Errorf("expected 48 but got %d", addr)
	}

	allocator.Free(64)
	allocator.Free(16)
	allocator.Free(48)
	addr = mustAlloc(allocator.Alloc(48))
	if addr != 16 {
		t.Errorf("expected 16 but got %d", addr)
	}
}

func TestBFCRandom(t *testing.T) {
	allocator := NewBFC(1<<20, 16)
	var used []usedChunk
	for i := 0; i < 10000; i++ {
		if len(used) > 0 && rand.Float64() < 0.6 {
			idx := rand.Intn(len(used))
			chunk := used[idx]
			used[idx] = used[len(used)-1]
			used = used[:len(used)-1]
			allocator.Free(chunk.start)
		} else {
			size := rand.Intn(1000)
			addr, err := allocator.Alloc(size)
			if err != nil {
				t.Fatal(err)
			}
			if addr >= (1 << 20) {
				t.Fatalf("address out of bounds: %d", addr)
			}
			for _, chunk := range used {
				if addr+size > chunk.start && addr < chunk.start+chunk.size {
					t.Fatal("memory chunk overlaps")
				}
			}
			used = append(used, usedChunk{start: addr, size: size})
		}
	}
	for _, x := range used {
		allocator.Free(x.start)
	}
	ptr := mustAlloc(allocator.Alloc(1<<20))
	if ptr != 0 {
		t.Errorf("expected 0 but got %d", ptr)
	}
}

func mustAlloc(addr int, err error) int {
	if err != nil {
		panic(err)
	}
	return addr
}

type usedChunk struct {
	start int
	size  int
}
