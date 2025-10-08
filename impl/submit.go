package impl

import (
	"context"

	"github.com/urfave/cli/v3"
)

// SubmitRequest
// LanguageID is defined by difference submit impletion
type SubmitRequest struct {
	Type       SubmitType
	LanguageID string
	FilePath   string
}

type SubmitType int

const (
	TypeSubmit = iota
	TypePrint
)

// SubmitResponse
// submitter should return callback url to enter frontend
// if any error is caught,ResultStatus is 1(StatusError)
// or set it 0(StatusOK)
type SubmitResponse struct {
	CallbackURL   string
	ResultID      string
	ResultStatue  Status
	ResultMessage string
}

type Status int

const (
	StatusOk = iota
	StatusError
)

// NewSubmitterFunc
// To initialize a new submitter
type NewSubmitterFunc func(...any) Submitter

// Submitter
// The receiver MUST BE Pointer Receiver
type Submitter interface {
	Before(context.Context, *SubmitRequest, *cli.Command) error
	Submit(context.Context, *SubmitRequest, *cli.Command) (*SubmitResponse, error)
	After(context.Context, *SubmitResponse, *cli.Command) error
}
