package util

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"strings"
)

const (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

type PasswordGenerator interface {
	Generate(passwordLength, minSpecialChar, minNum, minUpperCase int) string
}

func NewInPlacePasswordGenerator() PasswordGenerator {
	return &inPlacePasswordGenerator{}
}

type inPlacePasswordGenerator struct{}

func (*inPlacePasswordGenerator) Generate(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	var password strings.Builder

	// Set special character
	for i := 0; i < minSpecialChar; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet)))) //nolint:gosec

		password.WriteString(string(specialCharSet[random.Int64()]))
	}

	// Set numeric
	for i := 0; i < minNum; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet)))) //nolint:gosec

		password.WriteString(string(numberSet[random.Int64()]))
	}

	// Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet)))) //nolint:gosec

		password.WriteString(string(upperCharSet[random.Int64()]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase

	for i := 0; i < remainingLength; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet)))) //nolint:gosec

		password.WriteString(string(allCharSet[random.Int64()]))
	}

	inRune := []rune(password.String())

	mrand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}
