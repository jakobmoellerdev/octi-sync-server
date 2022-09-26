package memory

import (
	"context"
	"sync"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

func NewDevices() *Devices {
	return &Devices{sync.RWMutex{}, make(map[string]map[service.DeviceID]service.Device)}
}

type Devices struct {
	sync    sync.RWMutex
	devices map[string]map[service.DeviceID]service.Device
}

func (m *Devices) AddDevice(
	ctx context.Context, account service.Account, id service.DeviceID, password string,
) (service.Device, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	if m.devices[account.Username()] == nil {
		m.devices[account.Username()] = map[service.DeviceID]service.Device{}
	}

	m.devices[account.Username()][id] = service.NewBaseDevice(id, password)

	return m.devices[account.Username()][id], nil
}

func (m *Devices) GetDevices(
	_ context.Context, account service.Account,
) (map[service.DeviceID]service.Device, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	deviceIDs := m.devices[account.Username()]

	return deviceIDs, nil
}

func (m *Devices) GetDevice(_ context.Context, account service.Account, id service.DeviceID) (service.Device, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	return m.devices[account.Username()][id], nil
}

func (m *Devices) DeleteDevice(_ context.Context, account service.Account, id service.DeviceID) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	m.devices[account.Username()][id] = nil

	return nil
}

func (m *Devices) HealthCheck() service.HealthCheck {
	return func(_ context.Context) (string, bool) {
		return "memory-devices", true
	}
}
