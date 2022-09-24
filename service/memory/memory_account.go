package memory

type Account struct {
	username   string
	hashedPass string
}

func (m *Account) Username() string {
	return m.username
}

func (m *Account) HashedPass() string {
	return m.hashedPass
}

func NewAccount(username string, hashedPass string) *Account {
	return &Account{username: username, hashedPass: hashedPass}
}
