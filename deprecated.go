package main

import "fmt"

var (
	Deprecated = fmt.Errorf("This package is deprecated")
)

func main() {
	panic(Deprecated)
}
