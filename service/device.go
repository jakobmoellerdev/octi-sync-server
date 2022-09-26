package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"github.com/google/uuid"
)

//go:generate mockgen -source device.go -package mock -destination mock/device.go Device
type Device interface {
	ID() DeviceID
	Verify(password string) bool
	HashedPass() string
}

type DeviceID uuid.UUID

func (i DeviceID) String() string {
	return uuid.UUID(i).String()
}

func (i DeviceID) UUID() uuid.UUID {
	return uuid.UUID(i)
}

type BaseDevice struct {
	id         DeviceID
	hashedPass string
}

func (r *BaseDevice) ID() DeviceID {
	return r.id
}

func (r *BaseDevice) HashedPass() string {
	return r.hashedPass
}

func (r *BaseDevice) Verify(password string) bool {
	return subtle.ConstantTimeCompare([]byte(r.HashedPass()),
		[]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(password))))) == 1
}

func NewBaseDevice(deviceId DeviceID, hashedPass string) *BaseDevice {
	return &BaseDevice{deviceId, hashedPass}
}
