package redis

import "github.com/jakob-moeller-cloud/octi-sync-server/service"

type Device struct {
	id service.DeviceID
}

func (r *Device) ID() service.DeviceID {
	return r.id
}

func DeviceFromID(deviceId service.DeviceID) *Device {
	return &Device{deviceId}
}
