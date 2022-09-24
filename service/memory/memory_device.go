package memory

import (
	"github.com/google/uuid"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

type Device struct {
	id uuid.UUID
}

func (r *Device) ID() service.DeviceID {
	return service.DeviceID(r.id)
}

func DeviceFromID(deviceID service.DeviceID) *Device {
	return &Device{uuid.UUID(deviceID)}
}
