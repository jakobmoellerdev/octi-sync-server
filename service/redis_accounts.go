package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"octi-sync-server/middleware/auth/util"

	"github.com/go-redis/redis/v9"
)

const RedisAccountKeySpace = "octi:accounts"

func NewRedisAccounts(client *redis.Client) *RedisAccounts {
	return &RedisAccounts{client}
}

type RedisAccounts struct {
	client *redis.Client
}

func (r *RedisAccounts) Find(ctx context.Context, username string) (Account, error) {
	res, err := r.client.HGet(ctx, RedisAccountKeySpace, username).Result()
	if err != nil {
		return nil, err
	}
	return RedisAccountFromUsername(username, res), nil
}

func (r *RedisAccounts) FindHashed(ctx context.Context, hash string) (Account, error) {
	res, err := r.client.HGetAll(ctx, RedisAccountKeySpace).Result()
	if err != nil {
		return nil, err
	}

	for username, passHash := range res {
		if passHash == hash {
			return RedisAccountFromUsername(username, passHash), nil
		}
	}

	return nil, ErrAccountNotFound
}

func (r *RedisAccounts) Register(ctx context.Context, username string) (Account, string, error) {
	if acc, _ := r.Find(ctx, username); acc != nil {
		return nil, "", ErrAccountAlreadyExists
	}

	passLength, minSpecial, minNum, minUpper := 32, 6, 6, 6
	pass := util.NewInPlacePasswordGenerator().Generate(passLength, minSpecial, minNum, minUpper)
	hashedPass := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
	if err := r.client.HSet(ctx, RedisAccountKeySpace, username, hashedPass).Err(); err != nil {
		return nil, "", err
	}

	return RedisAccountFromUsername(username, hashedPass), pass, nil
}
