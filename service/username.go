package service

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/google/uuid"
)

//go:generate mockgen -source username.go -package mock -destination mock/username.go UsernameGenerator
type UsernameGenerator interface {
	Generate() (string, error)
}

type uuidUsernameGenerator struct {
	reader io.Reader
}

func (g *uuidUsernameGenerator) Generate() (string, error) {
	userID, err := uuid.NewRandomFromReader(g.reader)
	if err != nil {
		return "", fmt.Errorf("generating a uuid username for registration failed: %w", err)
	}

	return userID.String(), nil
}

type UsernameGenerationStrategy string

const UUIDUsernameGeneration UsernameGenerationStrategy = "uuid"

func NewUsernameGenerator(_ UsernameGenerationStrategy) (UsernameGenerator, error) {
	return &uuidUsernameGenerator{reader: rand.Reader}, nil
}
