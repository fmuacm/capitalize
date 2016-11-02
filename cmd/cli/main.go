package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fmuacm/capitalize"
)

func main() {
	if len(os.Args) == 1 {
		log.Println("Must pass in a string to capitalize.")
		os.Exit(0)
	}
	msg := strings.Join(os.Args[1:], " ")
	fmt.Println(capitalize.Format(msg))
}
