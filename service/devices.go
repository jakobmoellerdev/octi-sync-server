package service

import (
	"context"
	"errors"
)

//go:generate mockgen -source devices.go -package mock -destination mock/devices.go Devices
type Devices interface {
	AddDevice(ctx context.Context, account Account, id DeviceID, password string) (Device, error)
	GetDevices(ctx context.Context, account Account) (map[DeviceID]Device, error)
	GetDevice(ctx context.Context, account Account, id DeviceID) (Device, error)
	DeleteDevice(ctx context.Context, account Account, id DeviceID) error

	HealthCheck() HealthCheck
}

var ErrDeviceNotFound = errors.New("device not found")

func ErrIsDeviceNotFound(err error) bool {
	return errors.Is(err, ErrDeviceNotFound)
}