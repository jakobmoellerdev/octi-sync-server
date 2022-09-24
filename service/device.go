package service

import "github.com/google/uuid"

//go:generate mockgen -source device.go -package mock -destination mock/device.go Device
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

type BaseDevice struct {
	id DeviceID
}

func (r *BaseDevice) ID() DeviceID {
	return r.id
}

func DeviceFromID(deviceId DeviceID) *BaseDevice {
	return &BaseDevice{deviceId}
}
