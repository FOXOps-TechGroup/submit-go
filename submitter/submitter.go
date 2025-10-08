package submitter

import "github.com/FOXOps-TechGroup/submit-go/impl"

// NewSubmitterFunc
// to initialize submitter

var Submiiters map[string]impl.NewSubmitterFunc = map[string]impl.NewSubmitterFunc{}
