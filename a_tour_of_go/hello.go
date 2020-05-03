package main

import "fmt"

func main() {
	sl_a := []byte{'r', 'o', 'a', 'd'}
	sl_b := sl_a[1:2]
	fmt.Printf("len=%d cap=%d b: %s a: %s\n", len(sl_b), cap(sl_b), sl_b, sl_a)
	sl_b = append(sl_b, sl_a...)
	fmt.Printf("len=%d cap=%d b: %s a: %s\n", len(sl_b), cap(sl_b), sl_b, sl_a)
}
