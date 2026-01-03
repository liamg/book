package main

import (
	"fmt"
	"os"

	"github.com/liamg/book/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31mError: %s\x1b[0m\n", err)
	}
}
