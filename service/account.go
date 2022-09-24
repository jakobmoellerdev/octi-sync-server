package service

//go:generate mockgen -source account.go -package mock -destination mock/account.go Account
type Account interface {
	Username() string
	HashedPass() string
}

type BaseAccount struct {
	username   string
	hashedPass string
}

func (r *BaseAccount) Username() string {
	return r.username
}

func (r *BaseAccount) HashedPass() string {
	return r.hashedPass
}

func NewBaseAccount(username, hashedPass string) *BaseAccount {
	return &BaseAccount{username, hashedPass}
}
