package pkg

import (
	"crypto/rand"
	"math/big"
	"time"
)

const (
	SMSCodeLength     = 6
	SMSCodeExpireTime = 5 * time.Minute
)

func GenerateSMSCode() (string, error) {
	code := make([]byte, SMSCodeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte(num.Int64() + '0')
	}
	return string(code), nil
}
