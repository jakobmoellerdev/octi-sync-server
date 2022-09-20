package mem

import (
	"context"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

func NewDevices() *Devices {
	return &Devices{make(map[string][]string)}
}

type Devices struct {
	devices map[string][]string
}

func (m *Devices) FindByAccount(_ context.Context, acc service.Account) ([]service.Device, error) {
	deviceIDs := m.devices[acc.Username()]
	devices := make([]service.Device, len(deviceIDs))
	for i := range devices {
		devices[i] = DeviceFromID(deviceIDs[i])
	}
	return devices, nil
}

func (m *Devices) FindByDeviceID(_ context.Context, acc service.Account, deviceID string) (service.Device, error) {
	devices, noDeviceRegisteredYet := m.devices[acc.Username()]
	if noDeviceRegisteredYet {
		return nil, service.ErrDeviceNotFound
	}
	for i := range devices {
		if devices[i] == deviceID {
			return DeviceFromID(deviceID), nil
		}
	}
	return nil, service.ErrDeviceNotFound
}

func (m *Devices) Register(_ context.Context, acc service.Account, deviceID string) error {
	devices := m.devices[acc.Username()]
	m.devices[acc.Username()] = append(devices, deviceID)
	return nil
}

func (m *Devices) HealthCheck() service.HealthCheck {
	return func(_ context.Context) (string, bool) {
		return "memory-devices", true
	}
}
