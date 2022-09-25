package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
)

//go:generate mockgen -source account.go -package mock -destination mock/account.go Account
type Account interface {
	Username() string
	HashedPass() string
	Verify(password string) bool
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

func (r *BaseAccount) Verify(password string) bool {
	return subtle.ConstantTimeCompare([]byte(r.HashedPass()),
		[]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password))))) == 1
}
