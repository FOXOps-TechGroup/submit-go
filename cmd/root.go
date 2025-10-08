package cmd

import (
	"github.com/urfave/cli/v3"
)

var Root = &cli.Command{
	Name:    "submit",
	Usage:   "the script for XCPC submit",
	Flags:   flags,
	Version: "",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "FILE",
			UsageText: "The file to submit",
		},
	},
	// to ensure use any flags that have a single leading - or this will result in failures like "-version"
	// but "--version" is still valid
	UseShortOptionHandling: true,
	Before:                 BeforeHandle,
}
