package service

type RedisAccount struct {
	username   string
	hashedPass string
}

func (r *RedisAccount) Username() string {
	return r.username
}

func (r *RedisAccount) HashedPass() string {
	return r.hashedPass
}

func RedisAccountFromUsername(username string, hashedPass string) *RedisAccount {
	return &RedisAccount{username, hashedPass}
}
