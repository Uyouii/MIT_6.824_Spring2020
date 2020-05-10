package main

import (
	"encoding/json"
	"fmt"
)

type response2 struct {
	Page   int      `json:"page"`
	Fruits []string `json:"fruits"`
}

func main() {
	slcD := []string{"apple", "peach", "pear"}
	slcB, _ := json.Marshal(slcD)
	fmt.Println(string(slcB))

	str := `["apple","peach","pear"]`
	res := []string{}
	json.Unmarshal([]byte(str), &res)
	fmt.Println(res)
	fmt.Println(res[0])
}
