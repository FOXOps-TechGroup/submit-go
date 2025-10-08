package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/FOXOps-TechGroup/submit-go/cmd"
)

func main() {
	if err := cmd.Root.Run(
		context.Background(),
		os.Args,
	); err != nil {
		slog.Error(err.Error())
	}
}
