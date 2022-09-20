package redis

type Device struct {
	id string
}

func (r *Device) ID() string {
	return r.id
}

func DeviceFromID(deviceID string) *Device {
	return &Device{deviceID}
}
