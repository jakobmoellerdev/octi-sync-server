package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jakobmoellerdev/octi-sync-server/service"
)

func NewAccounts() *Accounts {
	return &Accounts{
		sync.RWMutex{},
		make(map[string][]byte),
		make(map[string][]string),
	}
}

type Accounts struct {
	sync     sync.RWMutex
	accounts map[string][]byte
	shares   map[string][]string
}

func (m *Accounts) Create(_ context.Context, username string) (service.Account, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	account := service.NewBaseAccount(username, time.Now())

	// cannot err out as time was created here
	createdAt, _ := account.CreatedAt().MarshalBinary()
	m.accounts[username] = createdAt

	return account, nil
}

func (m *Accounts) Find(_ context.Context, username string) (service.Account, error) {
	m.sync.RLock()
	defer m.sync.RUnlock()

	for user, createdAtRaw := range m.accounts {
		var createdAt time.Time
		if err := createdAt.UnmarshalBinary(createdAtRaw); err != nil {
			return nil, fmt.Errorf("error while parsing user creation: %w", err)
		}

		if user == username {
			return service.NewBaseAccount(user, createdAt), nil
		}
	}

	return nil, service.ErrAccountNotFound
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
