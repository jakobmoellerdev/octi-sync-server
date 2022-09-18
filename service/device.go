package service

import (
	"context"
	"errors"
)

type Device interface {
	ID() string
}

type Devices interface {
	FindByAccount(ctx context.Context, acc Account) ([]Device, error)
	FindByDeviceID(ctx context.Context, acc Account, deviceID string) (Device, error)
	Register(ctx context.Context, acc Account, deviceID string) error
}

var ErrDeviceNotFound = errors.New("device not found")
