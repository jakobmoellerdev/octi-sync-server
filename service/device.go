package service

//nolint:lll
//go:generate mockgen -package mock -destination mock/device.go github.com/jakob-moeller-cloud/octi-sync-server/service Device
type Device interface {
	ID() string
}
