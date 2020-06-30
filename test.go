package main

import (
	"ddz/utils"
	"fmt"
)

func main() {
	a := &lola{}
	b := &lola{A: 2, B: 3}
	utils.StructCopy(a, b)
	b.A = 5
	fmt.Println(a)
	fmt.Println(b)
}

type lola struct {
	A int
	B int
	C []int
}

func getData(a *lola) *lola {
	// a:= &lola{
	// 	A: 2,
	// 	B: 3,
	// 	C: []int{1, 2, 3},
	// }
	return a
}
