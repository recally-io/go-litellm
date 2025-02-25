package llms

import "errors"

var (
	ErrContentFieldsMisused = errors.New("can't use both Content and MultiContent properties simultaneously")
)
