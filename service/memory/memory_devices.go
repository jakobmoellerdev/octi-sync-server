package memory

import (
	"context"
	"crypto/sha256"
	"fmt"
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

func (r *Devices) hashPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func (r *Devices) AddDevice(
	_ context.Context, account service.Account, id service.DeviceID, password string,
) (service.Device, error) {
	r.sync.Lock()
	defer r.sync.Unlock()

	if r.devices[account.Username()] == nil {
		r.devices[account.Username()] = map[service.DeviceID]service.Device{}
	}

	r.devices[account.Username()][id] = service.NewBaseDevice(id, r.hashPassword(password))

	return r.devices[account.Username()][id], nil
}

func (r *Devices) GetDevices(
	_ context.Context, account service.Account,
) (map[service.DeviceID]service.Device, error) {
	r.sync.RLock()
	defer r.sync.RUnlock()

	deviceIDs := r.devices[account.Username()]

	return deviceIDs, nil
}

func (r *Devices) GetDevice(_ context.Context, account service.Account, id service.DeviceID) (service.Device, error) {
	r.sync.RLock()
	defer r.sync.RUnlock()

	device := r.devices[account.Username()][id]

	if device == nil {
		return nil, service.ErrDeviceNotFound
	}

	return device, nil
}

func (r *Devices) DeleteDevice(_ context.Context, account service.Account, id service.DeviceID) error {
	r.sync.Lock()
	defer r.sync.Unlock()

	r.devices[account.Username()][id] = nil

	return nil
}

func (r *Devices) HealthCheck() service.HealthCheck {
	return func(_ context.Context) (string, bool) {
		return "memory-devices", true
	}
}
