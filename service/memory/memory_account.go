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
