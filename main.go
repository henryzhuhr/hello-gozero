package main

import "time"

func main() {
	// 模拟一个任务的结束时间为29分钟前
	endTime := time.Now().Add(-29 * time.Minute)
	println(endTime.String())

	if time.Since(endTime) <= 30*time.Minute {
		println("任务在30分钟内完成")
	}

}
