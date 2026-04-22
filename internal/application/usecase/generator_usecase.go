package usecase

import (
	"crypto/rand"
	"math/big"
	"strings"
)

type GeneratorUseCase interface {
	Generate(length int, uppercase, lowercase, numbers, symbols bool) (string, error)
}

type generatorUseCase struct{}

func NewGeneratorUseCase() GeneratorUseCase {
	return &generatorUseCase{}
}

const (
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars    = "0123456789"
	symbolChars    = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

func (u *generatorUseCase) Generate(length int, uppercase, lowercase, numbers, symbols bool) (string, error) {
	if length < 8 {
		length = 8
	}
	if length > 64 {
		length = 64
	}

	var charset strings.Builder
	var required []byte

	if lowercase {
		charset.WriteString(lowercaseChars)
		char, _ := randomChar(lowercaseChars)
		required = append(required, char)
	}
	if uppercase {
		charset.WriteString(uppercaseChars)
		char, _ := randomChar(uppercaseChars)
		required = append(required, char)
	}
	if numbers {
		charset.WriteString(numberChars)
		char, _ := randomChar(numberChars)
		required = append(required, char)
	}
	if symbols {
		charset.WriteString(symbolChars)
		char, _ := randomChar(symbolChars)
		required = append(required, char)
	}

	charsetStr := charset.String()
	if charsetStr == "" {
		charsetStr = lowercaseChars + uppercaseChars + numberChars
		lowercase, uppercase, numbers = true, true, true
	}

	remaining := length - len(required)
	password := make([]byte, remaining)
	for i := range password {
		char, err := randomChar(charsetStr)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	result := append(required, password...)
	shuffle(result)

	return string(result), nil
}

func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func shuffle(b []byte) {
	for i := len(b) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := n.Int64()
		b[i], b[j] = b[j], b[i]
	}
}
