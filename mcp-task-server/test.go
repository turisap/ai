package main

import "fmt"

type a struct {
	s []string
}

func main() {
	str := a{s: []string{"a", "b"}}

	defer func(o a) {
		fmt.Println(len(o.s))
	}(str)

	str.s = nil
}
