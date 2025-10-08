package cmd

import (
	"context"
	"fmt"

	"github.com/FOXOps-TechGroup/submit-go/impl"
	"github.com/FOXOps-TechGroup/submit-go/utils"
	"github.com/urfave/cli/v3"
)

func CommandHandle(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("please input at least one file")
	}

	submitReq := &impl.SubmitRequest{}

	if cmd.Bool("print") {
		submitReq.Type = impl.TypePrint
	} else {
		submitReq.Type = impl.TypeSubmit
	}

	// Before submit
	// Submitter should complete the request
	// And prepare to submit
	err := currentSubmitter.Before(ctx, submitReq, cmd)
	if err != nil {
		return err
	}
	//here should wait to be ensured
	if utils.WaitToAccept() == false {
		ctx.Done()
		return nil
	}
	//submit this
	response, err := currentSubmitter.Submit(ctx, submitReq, cmd)
	if err != nil {
		return err
	}
	//after submit
	err = currentSubmitter.After(ctx, response, cmd)
	if err != nil {
		return err
	}
	return nil
}
