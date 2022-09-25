package memory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
)

func NewAccounts() *Accounts {
	return &Accounts{
		sync.RWMutex{},
		make(map[string]string),
		make(map[string][]string),
	}
}

type Accounts struct {
	sync     sync.RWMutex
	accounts map[string]string
	shares   map[string][]string
}

func (m *Accounts) Find(_ context.Context, username string) (service.Account, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	for user, pass := range m.accounts {
		if user == username {
			return service.NewBaseAccount(user, pass), nil
		}
	}

	return nil, service.ErrAccountNotFound
}

func (m *Accounts) Register(_ context.Context, username, password string) (service.Account, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	hashedPass := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	m.accounts[username] = hashedPass

	return service.NewBaseAccount(username, hashedPass), nil
}

func (m *Accounts) HealthCheck() service.HealthCheck {
	return func(_ context.Context) (string, bool) {
		return "memory-accounts", true
	}
}

func (m *Accounts) Share(_ context.Context, username string) (string, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	shares := m.shares[username]
	if shares == nil {
		shares = []string{}
	}

	shareUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error while generation share id for memory map: %w", err)
	}

	share := shareUUID.String()

	m.shares[username] = append(shares, share)

	return share, nil
}

func (m *Accounts) ActiveShares(_ context.Context, username string) ([]string, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	shares := m.shares[username]
	if shares == nil {
		shares = []string{}
	}

	return shares, nil
}

func (m *Accounts) IsShared(ctx context.Context, username string, share string) error {
	shares, _ := m.ActiveShares(ctx, username)

	for i := range shares {
		if shares[i] == share {
			return nil
		}
	}

	return service.ErrShareCodeInvalid
}

func (m *Accounts) Revoke(_ context.Context, username string, shareCode string) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	shares := m.shares[username]
	if shares == nil {
		shares = []string{}
	}

	for i := range shares {
		if shares[i] == shareCode {
			shares[i] = shares[len(shares)-1]
			shares = shares[:len(shares)-1]
			m.shares[username] = shares
		}
	}

	return nil
}
