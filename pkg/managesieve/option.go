package managesieve

import "time"

type Option interface {
	Set(params *parameters) error
}

type OptionFn func(params *parameters) error

func (fn OptionFn) Set(params *parameters) error {
	return fn(params)
}

func WithLogger(l Logger) OptionFn {
	return func(params *parameters) error {
		params.logger = l
		return nil
	}
}

func WithDialTimeout(t time.Duration) OptionFn {
	return func(params *parameters) error {
		params.dialTimeout = t
		return nil
	}
}
