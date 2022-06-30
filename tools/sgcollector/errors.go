package main

import "fmt"

// noVerboseError is an error that doesn't print the stack trace in zap.
type noVerboseError struct {
	err error
}

func (n noVerboseError) Unwrap() error {
	return n.err
}

func (n noVerboseError) Error() string {
	return n.err.Error()
}

// GenerateError reports that generation failed.
type GenerateError struct {
	stage   Stage
	notImpl []string
	err     error
}

func (p *GenerateError) Unwrap() error {
	return p.err
}

func (p *GenerateError) Error() string {
	return fmt.Sprintf("%s: %s", p.stage, p.err)
}
