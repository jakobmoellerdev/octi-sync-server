package service

import (
	"context"
	"errors"
)

//go:generate mockgen -source accounts.go -package mock -destination mock/accounts.go Accounts
type Accounts interface {
	Find(ctx context.Context, username string) (Account, error)
	Register(ctx context.Context, username string) (Account, string, error)
	Share(ctx context.Context, username string) (string, error)

	ActiveShares(ctx context.Context, username string) ([]string, error)
	IsShared(ctx context.Context, username string, share string) (bool, error)
	Revoke(ctx context.Context, username string, shareCode string) error

	HealthCheck() HealthCheck
}

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrAccountNotFound      = errors.New("account not found")
)
