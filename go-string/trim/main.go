package main

import (
	"fmt"
	"strings"
)

func main() {
	TrimRightDemo()
	TrimSuffixDemo()
	Cacul(125)
}

const a = 1 << (125 % 32)

type Set [int(8)]int32

func Cacul(c int) (s Set) {

	s[c/32] |= 1 << (c % 32)
	fmt.Printf("%v\n", s)
	return s
}

// TrimRightDemo 125 }
// 93 	]
func TrimRightDemo() {

	cutset := "好你"
	println(cutset[0])
	fmt.Println(strings.TrimRight("123你好你", cutset))
}

func TrimSuffixDemo() {
	fmt.Println(strings.TrimSuffix("123oxo", "xo"))
}
