package service

type RedisDevice struct {
	id string
}

func (r *RedisDevice) ID() string {
	return r.id
}

func RedisDeviceFromID(deviceID string) *RedisDevice {
	return &RedisDevice{deviceID}
}
