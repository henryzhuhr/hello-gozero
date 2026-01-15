// Package bloomfilter 布隆过滤器的例子
package bloomfilter

import (
	"fmt"
	"math/rand"

	"github.com/bits-and-blooms/bloom/v3"
)

// ExampleBasic demonstrates basic usage of BloomFilter with strings.
func ExampleBasic() {
	bitmapSize := uint64(1000) // 位图大小
	numHashes := uint64(4)     // 哈希函数数量
	userBF := NewBloomFilter[string](bitmapSize, numHashes, StringAdapter)

	usernamePrefix := "user_"

	maxUserCount := 1000
	usernames := make([]string, 0, maxUserCount)
	for i := 0; i < maxUserCount; i++ {
		username := fmt.Sprintf("%s%d", usernamePrefix, i)
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

// ExampleOfficialLibrary 官方库
func ExampleOfficialLibrary() {
	bitmapSize := uint(1000) // 位图大小
	numHashes := uint(4)     // 哈希函数数量
	bloomFilter := bloom.New(bitmapSize, numHashes)

	usernamePrefix := "user_"
	maxUserCount := 1000
	usernames := make([]string, 0, maxUserCount)
	for i := 0; i < maxUserCount; i++ {
		username := fmt.Sprintf("%s%d", usernamePrefix, i)
		bloomFilter.Add([]byte(username))
		usernames = append(usernames, username)
	}

	randomIndex := rand.Int31n(int32(maxUserCount))
	randomUsername := usernames[randomIndex]

	if bloomFilter.Test([]byte(randomUsername)) {
		fmt.Printf("Username %s may exist in the set.\n", randomUsername)
	} else {
		fmt.Printf("Username %s does not exist in the set.\n", randomUsername)
	}

}
