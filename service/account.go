package service

//nolint:lll
//go:generate mockgen -package mock -destination mock/account.go github.com/jakob-moeller-cloud/octi-sync-server/service Account
type Account interface {
	Username() string
	HashedPass() string
}
