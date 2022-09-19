package service

import (
	"context"
	"errors"
	"io"
)

type Module interface {
	Raw() io.Reader
	Size() int
}

type Modules interface {
	Set(ctx context.Context, name string, module Module) error
	Get(ctx context.Context, name string) (Module, error)
}

var (
	ErrWritingModuleFailed = errors.New("module write failed")
	ErrReadingModule       = errors.New("module read failed")
)
