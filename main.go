package main

import (
	"flag"
	"fmt"
)

func main() {
	var name string

	flag.StringVar(&name, "name", "", "Name")

	flag.Parse()

	fmt.Printf("Hello %s\n", name)
}
