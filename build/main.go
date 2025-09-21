package main

import (
	"fmt"
	"os"

	"github.com/FOXOps-TechGroup/submit-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
