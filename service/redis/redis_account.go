package redis

type Account struct {
	username   string
	hashedPass string
}

func (r *Account) Username() string {
	return r.username
}

func (r *Account) HashedPass() string {
	return r.hashedPass
}

func AccountFromUsername(username, hashedPass string) *Account {
	return &Account{username, hashedPass}
}
