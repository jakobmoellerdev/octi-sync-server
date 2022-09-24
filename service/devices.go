package service

import (
	"context"
	"errors"
)

//go:generate mockgen -source devices.go -package mock -destination mock/devices.go Devices
type Devices interface {
	FindByAccount(ctx context.Context, acc Account) ([]Device, error)
	FindByDeviceID(ctx context.Context, acc Account, deviceId DeviceID) (Device, error)
	Register(ctx context.Context, acc Account, deviceId DeviceID) (Device, error)
	HealthCheck() HealthCheck
}

var ErrDeviceNotFound = errors.New("device not found")
