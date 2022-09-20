package mem

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
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
			return &Account{username: username, hashedPass: pass}, nil
		}
	}
	return nil, service.ErrAccountNotFound
}

func (m *Accounts) Register(_ context.Context, username string) (service.Account, string, error) {
	m.sync.Lock()
	defer m.sync.Unlock()
	//nolint:gomnd
	pass := util.NewInPlacePasswordGenerator().Generate(9, 3, 3, 3)
	hashedPass := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
	m.accounts[username] = fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
	return &Account{username: username, hashedPass: hashedPass}, pass, nil
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
		return "", err
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

func (m *Accounts) IsShared(ctx context.Context, username string, share string) (bool, error) {
	shares, _ := m.ActiveShares(ctx, username)
	for i := range shares {
		if shares[i] == share {
			return true, nil
		}
	}
	return false, nil
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
