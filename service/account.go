package service

import "time"

//go:generate mockgen -source account.go -package mock -destination mock/account.go Account
type Account interface {
	Username() string
	CreatedAt() time.Time
}

type BaseAccount struct {
	username  string
	createdAt time.Time
}

func (r *BaseAccount) Username() string {
	return r.username
}

func (r *BaseAccount) CreatedAt() time.Time {
	return r.createdAt
}

func NewBaseAccount(username string, createdAt time.Time) *BaseAccount {
	return &BaseAccount{username, createdAt}
}