package utils

import (
	"math/rand/v2"
	"strings"
)

const LETTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(length int) string {
	var sb strings.Builder

	sb.WriteString("mw_")

	for range length {
		sb.WriteByte(LETTERS[rand.IntN(len(LETTERS))])
	}

	return sb.String()
}
