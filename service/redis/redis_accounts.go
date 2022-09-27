package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
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

func (r *Accounts) shareKey(account service.Account) string {
	return fmt.Sprintf("%s:%s", ShareKeySpace, account.Username())
}

func (r *Accounts) Share(ctx context.Context, account service.Account) (string, error) {
	shareID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error during share code generation: %w", err)
	}

	shareCode := shareID.String()
	if err := r.Client.LPush(ctx, r.shareKey(account), shareCode).Err(); err != nil {
		return "", fmt.Errorf("error while pushing shareCode: %w", err)
	}

	if err := r.Client.Expire(ctx, r.shareKey(account), time.Hour).Err(); err != nil {
		return "", fmt.Errorf("error while setting expiry: %w", err)
	}

	return shareCode, nil
}

func (r *Accounts) ActiveShares(ctx context.Context, account service.Account) ([]string, error) {
	shareCodes, err := r.Client.LRange(ctx, r.shareKey(account), 0, -1).Result()
	if err != nil {
		return []string{}, fmt.Errorf("could not determine active sharecode list: %w", err)
	}

	return shareCodes, nil
}

func (r *Accounts) IsShared(ctx context.Context, account service.Account, shareCode service.ShareCode) error {
	err := r.Client.LPos(ctx, r.shareKey(account), shareCode.String(), redis.LPosArgs{}).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return service.ErrShareCodeInvalid
		}

		return fmt.Errorf("could not find out if share code is valid: %w", err)
	}

	return nil
}

func (r *Accounts) Revoke(ctx context.Context, account service.Account, shareCode service.ShareCode) error {
	if err := r.Client.LRem(ctx, r.shareKey(account), 0, shareCode).Err(); err != nil {
		return fmt.Errorf("error while revoking user: %w", err)
	}

	return nil
}
