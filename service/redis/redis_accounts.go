package redis

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"

	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/util"
)

const (
	AccountKeySpace = "octi:accounts"
	ShareKeySpace   = "octi:accounts:share"
)

func NewAccounts(client *redis.Client) *Accounts {
	return &Accounts{client}
}

type Accounts struct {
	client *redis.Client
}

func (r *Accounts) Find(ctx context.Context, username string) (service.Account, error) {
	res, err := r.client.HGet(ctx, AccountKeySpace, username).Result()
	if err != nil {
		return nil, fmt.Errorf("error while looking up user: %w", err)
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
		return nil, "", fmt.Errorf("error while setting user in account key space: %w", err)
	}

	return AccountFromUsername(username, hashedPass), pass, nil
}

func (r *Accounts) HealthCheck() service.HealthCheck {
	return func(ctx context.Context) (string, bool) {
		return "redis-accounts", r.client.Ping(ctx).Err() == nil
	}
}

func (r *Accounts) Share(ctx context.Context, username string) (string, error) {
	shareID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error during share code generation: %w", err)
	}

	shareCode := shareID.String()
	if err := r.client.WithTimeout(1*time.Hour).LPush(ctx, r.shareKey(username), shareCode).Err(); err != nil {
		return "", fmt.Errorf("error while pushing shareCode: %w", err)
	}

	return shareCode, nil
}

func (r *Accounts) ActiveShares(ctx context.Context, username string) ([]string, error) {
	shareCodes, err := r.client.LRange(ctx, r.shareKey(username), 0, -1).Result()
	if err != nil {
		return []string{}, fmt.Errorf("could not determine active sharecode list: %w", err)
	}

	return shareCodes, nil
}

func (r *Accounts) IsShared(ctx context.Context, username string, share string) (bool, error) {
	err := r.client.LPos(ctx, r.shareKey(username), share, redis.LPosArgs{}).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, fmt.Errorf("could not find out if share code is valid: %w", err)
	}

	return true, nil
}

func (r *Accounts) Revoke(ctx context.Context, username string, shareCode string) error {
	if err := r.client.LRem(ctx, r.shareKey(username), 0, shareCode).Err(); err != nil {
		return fmt.Errorf("error while revoking user: %w", err)
	}

	return nil
}

func (r *Accounts) shareKey(username string) string {
	return fmt.Sprintf("%s:%s", ShareKeySpace, username)
}
