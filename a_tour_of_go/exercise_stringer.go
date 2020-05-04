package main

import "fmt"

type IPAddr [4]byte

func (ipAddr IPAddr) String() string {
	var ipStr string
	for index, num := range ipAddr {
		ipStr += fmt.Sprintf("%d", num)
		if index != len(ipAddr)-1 {
			ipStr += "."
		}
	}
	return ipStr
}

func main() {
	hosts := map[string]IPAddr{
		"loopback":  {127, 0, 0, 1},
		"googleDNS": {8, 8, 8, 8},
	}
	for name, ip := range hosts {
		fmt.Printf("%v: %v\n", name, ip)
	}
}
