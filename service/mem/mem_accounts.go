package mem

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
)

func NewAccounts() *Accounts {
	return &Accounts{make(map[string]string)}
}

type Accounts struct {
	accounts map[string]string
}

func (m *Accounts) Find(_ context.Context, username string) (service.Account, error) {
	for user, pass := range m.accounts {
		if user == username {
			return &Account{username: username, hashedPass: pass}, nil
		}
	}
	return nil, service.ErrAccountNotFound
}

func (m *Accounts) Register(_ context.Context, username string) (service.Account, string, error) {
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
