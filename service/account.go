package service

import (
	"context"
	"errors"
)

type Account interface {
	Username() string
	HashedPass() string
}

type Accounts interface {
	Find(ctx context.Context, username string) (Account, error)
	FindHashed(ctx context.Context, hash string) (Account, error)
	Register(ctx context.Context, username string) (Account, string, error)
}

var ErrAccountAlreadyExists = errors.New("account already exists")
var ErrAccountNotFound = errors.New("account not found")
