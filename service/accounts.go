package service

import (
	"context"
	"errors"
)

//go:generate mockgen -source accounts.go -package mock -destination mock/accounts.go Accounts
type Accounts interface {
	Find(ctx context.Context, username string) (Account, error)
	Create(ctx context.Context, username string) (Account, error)

	HealthCheck() HealthCheck
}

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrAccountNotFound      = errors.New("account not found")
	ErrShareCodeInvalid     = errors.New("share code is invalid")
)
