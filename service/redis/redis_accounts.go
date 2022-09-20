package redis

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"

	"github.com/go-redis/redis/v9"
)

const AccountKeySpace = "octi:accounts"

func NewAccounts(client *redis.Client) *Accounts {
	return &Accounts{client}
}

type Accounts struct {
	client *redis.Client
}

func (r *Accounts) Find(ctx context.Context, username string) (service.Account, error) {
	res, err := r.client.HGet(ctx, AccountKeySpace, username).Result()
	if err != nil {
		return nil, err
	}
	return AccountFromUsername(username, res), nil
}

func (r *Accounts) Register(ctx context.Context, username string) (service.Account, string, error) {
	if acc, _ := r.Find(ctx, username); acc != nil {
		return nil, "", service.ErrAccountAlreadyExists
	}

	passLength, minSpecial, minNum, minUpper := 32, 6, 6, 6
	pass := util.NewInPlacePasswordGenerator().Generate(passLength, minSpecial, minNum, minUpper)
	hashedPass := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
	if err := r.client.HSet(ctx, AccountKeySpace, username, hashedPass).Err(); err != nil {
		return nil, "", err
	}

	return AccountFromUsername(username, hashedPass), pass, nil
}

func (r *Accounts) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-accounts", r.client.Ping(ctx).Err() == nil
	}
}
