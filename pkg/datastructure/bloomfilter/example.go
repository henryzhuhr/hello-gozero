// Package bloomfilter 布隆过滤器的例子
package bloomfilter

import (
	"fmt"
	"math/rand"
)

// ExampleBasic demonstrates basic usage of BloomFilter with strings.
func ExampleBasic() {
	bitmapSize := uint64(1000) // 位图大小
	numHashes := uint64(4)     // 哈希函数数量
	userBF := NewBloomFilter[string](bitmapSize, numHashes, StringAdapter)

	useranamePrefix := "user_"

	maxUserCount := 1000
	usernames := make([]string, 0, maxUserCount)
	for i := 0; i < maxUserCount; i++ {
		username := fmt.Sprintf("%s%d", useranamePrefix, i)
		userBF.Add(username)
		usernames = append(usernames, username)
	}

	randomIndex := rand.Int31n(int32(maxUserCount))

	randomUsername := usernames[randomIndex]

	if userBF.Contains(randomUsername) {
		fmt.Printf("Username %s may exist in the set.\n", randomUsername)
	} else {
		fmt.Printf("Username %s does not exist in the set.\n", randomUsername)
	}
}
