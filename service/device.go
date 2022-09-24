package service

import "github.com/google/uuid"

//nolint:lll
//go:generate mockgen -package mock -destination mock/device.go github.com/jakob-moeller-cloud/octi-sync-server/service Device
type Device interface {
	ID() DeviceID
}

type DeviceID uuid.UUID

func (i DeviceID) String() string {
	return uuid.UUID(i).String()
}

func (i DeviceID) UUID() uuid.UUID {
	return uuid.UUID(i)
}
