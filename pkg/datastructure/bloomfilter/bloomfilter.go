// Package bloomfilter 布隆过滤器
package bloomfilter

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math/bits"
)

// ---------------------- 泛型类型转换适配器 ----------------------

// HashAdapter 泛型类型转换适配器，用于将任意类型 T 转换为 []byte 字节流
type HashAdapter[T any] func(data T) []byte

// StringAdapter string 类型的默认适配器（最常用）
// 直接将 string 转换为 []byte，高效且无额外开销
var StringAdapter HashAdapter[string] = func(data string) []byte {
	return []byte(data)
}

// IntAdapter int 类型的默认适配器（适配 32/64 位系统的 int 类型）
// 将 int 转换为 8 字节 []byte（大端序，保证跨系统一致性）
var IntAdapter HashAdapter[int] = func(data int) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(data))
	return buf
}

// Uint64Adapter uint64 类型的默认适配器
// 常用于 ID 类字段（如用户 ID、商品 ID），跨系统一致性好
var Uint64Adapter HashAdapter[uint64] = func(data uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, data)
	return buf
}

// ByteSliceAdapter []byte 类型的默认适配器
// 直接返回原切片，避免冗余拷贝
var ByteSliceAdapter HashAdapter[[]byte] = func(data []byte) []byte {
	return data
}

// ---------------------- 布隆过滤器 ----------------------

// BloomFilter 布隆过滤器
type BloomFilter[T any] interface {
	// Add 添加元素到布隆过滤器
	Add(data T)

	// Contains 检查元素是否可能存在于布隆过滤器中
	// 返回 true：大概率存在（可能假阳性）；返回 false：一定不存在（无假阴性）
	Contains(data T) bool
}

// basicBloomFilter 布隆过滤器
type basicBloomFilter[T any] struct {
	bitmap      []uint64       // 位图存储（用uint64切片，每一位表示一个标记，节省内存）
	bitSize     uint64         // 位图总比特数
	numBits     uint64         // 位数组总位数（m）
	numHashes   uint64         // 哈希函数个数（k）
	hasher      hash.Hash64    // 基础哈希函数（用于生成两个独立哈希）
	hashAdapter HashAdapter[T] // 类型转 []byte 的适配器
}

// NewBloomFilter 创建布隆过滤器
//
//   - size: 		位数组的总位数（m），必须 > 0
//   - k: 			哈希函数个数（k），建议 1 <= k <= 10（通常 3~7）
//   - hashAdapter: 将 T 转为 []byte 的函数
func NewBloomFilter[T any](size uint64, k uint64, hashAdapter HashAdapter[T]) BloomFilter[T] {
	if size == 0 {
		size = 1
	}
	if k == 0 {
		k = 1
	}
	// 限制 k 避免溢出或性能问题（可选）
	if k > 32 {
		k = 32
	}

	// 参数合法性校验
	if size == 0 {
		panic("bloomfilter: size 不能为 0（位图总比特数必须大于 0）")
	}
	if k == 0 {
		panic("bloomfilter: k 不能为 0（哈希函数个数必须大于 0）")
	}
	if hashAdapter == nil {
		panic("bloomfilter: hashAdapter 不能为 nil（必须提供泛型类型转换适配器）")
	}

	// 计算 uint64 切片的长度（向上取整，保证能容纳 size 个比特位）
	sliceLen := (uint64(size) + 63) / 64
	bitmap := make([]uint64, sliceLen)

	// 默认使用 fnv-1a 64位哈希函数（高效、分布均匀，可替换为其他 hash.Hash64 实现）
	hasher := fnv.New64a()

	return &basicBloomFilter[T]{
		bitmap:      bitmap,
		bitSize:     uint64(size),
		numBits:     size,
		numHashes:   k,
		hasher:      hasher,
		hashAdapter: hashAdapter,
	}
}

// Add Implements [BloomFilter.Add]
func (bf *basicBloomFilter[T]) Add(data T) {
	// 1. 通过适配器将泛型 T 转换为 []byte（解决任意类型哈希问题）
	dataBytes := bf.hashAdapter(data)

	// 2. 生成多个离散的哈希索引（基于单个 hash.Hash64 模拟多个独立哈希函数）
	indexes := bf.generateHashIndexes(dataBytes)

	// 3. 遍历所有索引，将对应位图的比特位置为 1（位运算高效操作）
	for _, idx := range indexes {
		// 计算：索引对应的 uint64 切片下标 + 对应的比特位下标
		sliceIdx := idx / 64
		bitIdx := idx % 64

		// 按位或操作，将对应比特位置为 1（不影响其他比特位的状态）
		bf.bitmap[sliceIdx] |= 1 << bitIdx
	}
}

// Contains Implements [BloomFilter.Contains]
func (bf *basicBloomFilter[T]) Contains(data T) bool {
	// 1. 通过适配器将泛型 T 转换为 []byte
	dataBytes := bf.hashAdapter(data)

	// 2. 生成多个离散的哈希索引
	indexes := bf.generateHashIndexes(dataBytes)

	// 3. 遍历所有索引，检查对应位图的比特位是否全部为 1
	for _, idx := range indexes {
		sliceIdx := idx / 64
		bitIdx := idx % 64

		// 按位与操作，判断对应比特位是否为 1（为 0 则直接返回 false，一定不存在）
		if (bf.bitmap[sliceIdx] & (1 << bitIdx)) == 0 {
			return false
		}
	}

	// 所有比特位都为 1，大概率存在（可能存在假阳性，布隆过滤器正常特性）
	return true
}

// generateHashIndexes 生成多个离散的哈希索引（核心辅助方法，非导出）
// 基于 Kirsch-Mitzenmacher 优化算法：用两个独立哈希值生成 k 个哈希索引，高效且分布均匀
// 输入：[]byte 字节流（已通过 HashAdapter 转换）
// 输出：k 个不越界的位图索引
func (bf *basicBloomFilter[T]) generateHashIndexes(data []byte) []uint64 {
	// 重置哈希函数状态（避免累计哈希值，保证每次计算独立）
	bf.hasher.Reset()

	// 写入数据并计算第一个 64 位哈希值 h1
	_, _ = bf.hasher.Write(data)
	h1 := bf.hasher.Sum64()

	// 再次重置并写入数据，生成第二个 64 位哈希值 h2（模拟第二个独立哈希函数）
	bf.hasher.Reset()
	_, _ = bf.hasher.Write(data)
	h2 := bits.Reverse64(h1) // 反转比特位生成独立的 h2，替代额外引入哈希函数

	// 生成 numHashes 个离散索引
	indexes := make([]uint64, bf.numHashes)
	for i := uint64(0); i < bf.numHashes; i++ {
		// Kirsch-Mitzenmacher 算法：h = h1 + i * h2，保证索引离散性
		hashValue := h1 + uint64(i)*h2
		// 取模保证索引在 bitmap 比特范围内，不越界
		indexes[i] = hashValue % bf.bitSize
	}

	return indexes
}
