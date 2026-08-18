package uuid

import (
	"math/rand"

	"github.com/google/uuid"
)

const (
	namePrefix = "sensecraftVoice-"
)

func NewUUID() string {
	return uuid.New().String()
}

func NewRandName(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")

	rs := make([]rune, length)
	for i := 0; i < length; i++ {
		rs[i] = chars[rand.Intn(len(chars))]
	}

	return namePrefix + string(rs)
}
