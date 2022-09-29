package service

import "context"

type ShareCode string

func (c ShareCode) String() string {
	return string(c)
}

//go:generate mockgen -source sharing.go -package mock -destination mock/sharing.go Sharing
type Sharing interface {
	Share(ctx context.Context, account Account) (ShareCode, error)
	Shared(ctx context.Context, shareCode ShareCode) (Account, error)
	Revoke(ctx context.Context, shareCode ShareCode) error
}
