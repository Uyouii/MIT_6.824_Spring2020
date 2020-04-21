package main

import (
	"fmt"
	"math"
)

func Sqrt(x float64) float64 {
	var z float64 = x / 2

	oldZ := 0.0

	for math.Abs(z-oldZ) >= 0.0001 {
		oldZ = z
		z -= (z*z - x) / (2 * z)
		// fmt.Println(z)
	}

	return z
}

func main() {
	fmt.Println(Sqrt(2))
	fmt.Println(math.Sqrt(2))
}
