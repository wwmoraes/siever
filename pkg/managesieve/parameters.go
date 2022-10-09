package managesieve

import (
	"io"
	"log"
	"time"
)

type parameters struct {
	logger      Logger
	dialTimeout time.Duration
}

func newParameters(options ...Option) (params *parameters, err error) {
	params = &parameters{
		dialTimeout: 5 * time.Second,
	}

	for _, option := range options {
		err = option.Set(params)
		if err != nil {
			return nil, err
		}
	}

	// we create the default logger after the options are processed to prevent an
	// useless allocation
	if params.logger == nil {
		params.logger = log.New(io.Discard, "DEBUG: ", 0)
	}

	return params, err
}
