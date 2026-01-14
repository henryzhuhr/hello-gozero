package bloomfilter

import (
	"fmt"
	"testing"
)

func TestBasicAddContainsString(t *testing.T) {
	bf := NewBloomFilter(10000, 4, StringAdapter)
	items := []string{"hello", "world", "foo", "bar"}
	for _, it := range items {
		bf.Add(it)
	}

	for _, it := range items {
		if !bf.Contains(it) {
			t.Fatalf("expected bloom filter to contain %s", it)
		}
	}

	if bf.Contains("not-present-value-xyz") {
		t.Fatalf("expected bloom filter to NOT contain a value that was not added")
	}
}

func TestAdapters_IntAndUint64(t *testing.T) {
	bfInt := NewBloomFilter[int](8192, 3, IntAdapter)
	bfInt.Add(42)
	if !bfInt.Contains(42) {
		t.Fatal("expected int adapter bloom filter to contain 42")
	}

	bfU := NewBloomFilter[uint64](8192, 3, Uint64Adapter)
	bfU.Add(1234567890)
	if !bfU.Contains(1234567890) {
		t.Fatal("expected uint64 adapter bloom filter to contain 1234567890")
	}
}

func TestGenerateHashIndexesBounds(t *testing.T) {
	bf := NewBloomFilter[string](100, 5, StringAdapter)
	// type assert to access internal fields for stronger invariants
	bfp := bf.(*basicBloomFilter[string])
	idxs := bfp.generateHashIndexes([]byte("test-indexes"))
	if len(idxs) != int(bfp.numHashes) {
		t.Fatalf("expected %d indexes, got %d", bfp.numHashes, len(idxs))
	}
	for _, idx := range idxs {
		if idx >= bfp.bitSize {
			t.Fatalf("index out of bounds: %d >= %d", idx, bfp.bitSize)
		}
	}
}

func TestNewBloomFilter_NilAdapterPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected NewBloomFilter to panic when hashAdapter is nil")
		}
	}()
	_ = NewBloomFilter[string](10, 1, nil)
}

func BenchmarkAddString(b *testing.B) {
	bf := NewBloomFilter[string](100000, 4, StringAdapter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(fmt.Sprintf("val-%d", i))
	}
}

func BenchmarkContainsStringHit(b *testing.B) {
	bf := NewBloomFilter[string](100000, 4, StringAdapter)
	n := 10000
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("k-%d", i)
		bf.Add(keys[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.Contains(keys[i%len(keys)])
	}
}

func BenchmarkContainsStringMiss(b *testing.B) {
	bf := NewBloomFilter[string](100000, 4, StringAdapter)
	n := 10000
	for i := 0; i < n; i++ {
		bf.Add(fmt.Sprintf("k-%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.Contains(fmt.Sprintf("not-exist-%d", i))
	}
}

func BenchmarkGenerateHashIndexes(b *testing.B) {
	bf := NewBloomFilter[string](100000, 4, StringAdapter)
	bfp := bf.(*basicBloomFilter[string])
	data := []byte("benchmark-data")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bfp.generateHashIndexes(data)
	}
}
