package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func Int32Ptr(i int32) *int32 { return &i }

func GenerateUniqueID() string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(randomBytes)
}
