package service

import "context"

type ShareCode string

func (c ShareCode) String() string {
	return string(c)
}

//go:generate mockgen -source sharing.go -package mock -destination mock/sharing.go Sharing
type Sharing interface {
	Share(ctx context.Context, account Account) (string, error)
	ActiveShares(ctx context.Context, account Account) ([]string, error)
	IsShared(ctx context.Context, account Account, shareCode ShareCode) error
	Revoke(ctx context.Context, account Account, shareCode ShareCode) error
}
