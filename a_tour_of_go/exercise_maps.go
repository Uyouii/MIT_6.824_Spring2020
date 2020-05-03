package main

import (
	"strings"

	"golang.org/x/tour/wc"
)

func WordCount(s string) map[string]int {
	res := make(map[string]int)
	str_list := strings.Fields(s)
	for _, value := range str_list {
		res[value] += 1
	}
	return res
}

func main() {
	wc.Test(WordCount)
}
