package service

import (
	"context"
	"errors"
)

//nolint:lll
//go:generate mockgen -package mock -destination mock/modules.go github.com/jakob-moeller-cloud/octi-sync-server/service Modules
type Modules interface {
	Set(ctx context.Context, name string, module Module) error
	Get(ctx context.Context, name string) (Module, error)
	HealthCheck() HealthCheck
}

var (
	ErrWritingModuleFailed = errors.New("module write failed")
	ErrReadingModule       = errors.New("module read failed")
)
