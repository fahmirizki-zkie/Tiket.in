package helper

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func GenerateTicketCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "TIX-" + strings.ToUpper(hex.EncodeToString(b)), nil
}
