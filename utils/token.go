// utils/token.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken menghasilkan string token acak dengan panjang n byte (hasilnya 2n hex characters)
func GenerateRandomToken(n int) (string, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
