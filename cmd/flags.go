package cmd

import (
	"github.com/urfave/cli/v3"
)

var flags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "to get verbose output",
	},
	&cli.BoolFlag{
		Name:    "print",
		Aliases: []string{"p"},
		Usage:   "to submit a print request",
	},
	&cli.BoolFlag{
		Name:  "version",
		Usage: "check the version",
	},
	&cli.StringFlag{
		Name:    "cid",
		Aliases: []string{"c"},
		Usage:   "submit to this competition",
	},
	&cli.StringFlag{
		Name:    "problem",
		Aliases: []string{"pid"},
		Usage:   "submit to this problem",
	},
}
