package memory

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"sync"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

func NewModules() *Modules {
	return &Modules{sync.RWMutex{}, make(map[string][]byte)}
}

type Modules struct {
	sync sync.RWMutex
	data map[string][]byte
}

func (m *Modules) DeleteByPattern(_ context.Context, pattern string) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	for key := range m.data {
		if matched, err := regexp.Match(pattern, []byte(key)); matched {
			m.data[key] = nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (m *Modules) Set(_ context.Context, name string, module service.Module) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	moduleData, err := io.ReadAll(module.Raw())
	if err != nil {
		return fmt.Errorf("error while reading module raw input for writing: %w", err)
	}

	m.data[name] = moduleData

	return nil
}

func (m *Modules) Get(_ context.Context, name string) (service.Module, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	return ModuleFromBytes(m.data[name]), nil
}

func (m *Modules) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "memory-modules", true
	}
}
