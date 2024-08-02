package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/jakobmoellerdev/octi-sync-server/service"
)

const (
	AccountKeySpace = "octi:accounts"
	ShareKeySpace   = "octi:accounts:share"
)

type Accounts struct {
	Client redis.Cmdable
}

func (r *Accounts) Create(ctx context.Context, username string) (service.Account, error) {
	if acc, _ := r.Find(ctx, username); acc != nil {
		return nil, service.ErrAccountAlreadyExists
	}

	account := service.NewBaseAccount(username, time.Now())

	// cannot err out as time was created here
	createdAt, _ := account.CreatedAt().MarshalBinary()

	if err := r.Client.HSet(ctx, AccountKeySpace, username, createdAt).Err(); err != nil {
		return nil, fmt.Errorf("error while setting user in account key space: %w", err)
	}

	return account, nil
}

func (r *Accounts) Find(ctx context.Context, username string) (service.Account, error) {
	res, err := r.Client.HGet(ctx, AccountKeySpace, username).Result()

	if err == redis.Nil {
		return nil, service.ErrAccountNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("error while looking up user: %w", err)
	}

	var createdAt time.Time
	if err = createdAt.UnmarshalBinary([]byte(res)); err != nil {
		return nil, fmt.Errorf("error while parsing user creation: %w", err)
	}

	return service.NewBaseAccount(username, createdAt), nil
}

func (r *Accounts) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-accounts", r.Client.Ping(ctx).Err() == nil
	}
}

func (r *Accounts) shareKey(shareCode service.ShareCode) string {
	return fmt.Sprintf("%s:%s", ShareKeySpace, shareCode.String())
}

func (r *Accounts) Share(ctx context.Context, account service.Account) (service.ShareCode, error) {
	shareID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error during share code generation: %w", err)
	}

	shareCode := service.ShareCode(shareID.String())
	if err := r.Client.Set(
		ctx,
		r.shareKey(shareCode),
		account.Username(),
		time.Hour,
	).Err(); err != nil {
		return "", fmt.Errorf("error while pushing shareCode: %w", err)
	}

	return shareCode, nil
}

func (r *Accounts) Shared(ctx context.Context, shareCode service.ShareCode) (service.Account, error) {
	accountID, err := r.Client.Get(ctx, r.shareKey(shareCode)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, service.ErrShareCodeInvalid
		}

		return nil, fmt.Errorf("could not find out if share code is valid: %w", err)
	}

	return r.Find(ctx, accountID)
}

func (r *Accounts) Revoke(ctx context.Context, shareCode service.ShareCode) error {
	if err := r.Client.Del(ctx, r.shareKey(shareCode)).Err(); err != nil {
		return fmt.Errorf("error while revoking share code: %w", err)
	}

	return nil
}
