package utils

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func WaitToAccept() bool {

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Continue? [y/N]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "n" {
			return false
		} else if input == "y" {
			return true
		}

		fmt.Println("Please type 'y/Y' or 'n/N'")
	}
}

func InitLog(level slog.Level) {
	Logger = slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     level,
			},
		),
	)
}
