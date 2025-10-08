package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/FOXOps-TechGroup/submit-go/config"
	"github.com/FOXOps-TechGroup/submit-go/impl"
	"github.com/FOXOps-TechGroup/submit-go/submitter"
	"github.com/FOXOps-TechGroup/submit-go/utils"
	"github.com/urfave/cli/v3"
)

var currentSubmitter impl.Submitter

func BeforeHandle(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	//Log init
	if cmd.Bool("verbose") {
		utils.InitLog(slog.LevelDebug)
	} else {
		utils.InitLog(slog.LevelInfo)
	}

	//read config
	err := config.Read()
	if err != nil {
		if errors.Is(err, config.EditSettingsError{}) {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		return ctx, err
	}

	//set submitter
	sf, ok := submitter.Submiiters[config.GlobalConfig.Submitter]
	if ok {
		currentSubmitter = sf()
	} else {
		return ctx, fmt.Errorf("can not find submitter: %s", config.GlobalConfig.Submitter)
	}

	return ctx, nil
}
