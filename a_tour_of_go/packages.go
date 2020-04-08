package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	iNowTime := time.Now().UnixNano()
	fmt.Println("nowtime:", iNowTime)
	rand.Seed(iNowTime)
	fmt.Printf("my favorite number is %v", rand.Intn(10000))
}
