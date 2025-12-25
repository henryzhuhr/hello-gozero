package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Duration(5) * time.Second)
	fmt.Println(time.Duration(-5) * time.Second)
}
