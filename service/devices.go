package service

import (
	"context"
	"errors"
)

//nolint:lll
//go:generate mockgen -package mock -destination mock/devices.go github.com/jakob-moeller-cloud/octi-sync-server/service Devices
type Devices interface {
	FindByAccount(ctx context.Context, acc Account) ([]Device, error)
	FindByDeviceID(ctx context.Context, acc Account, deviceID string) (Device, error)
	Register(ctx context.Context, acc Account, deviceID string) (Device, error)
	HealthCheck() HealthCheck
}

var ErrDeviceNotFound = errors.New("device not found")
