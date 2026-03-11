package crypto

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// GenerateSHA3_256 menghasilkan hash SHA3-256 dari string input
func GenerateSHA3_256(data string) string {
	// sha3.Sum256 menghasilkan array byte berukuran 32
	hash := sha3.Sum256([]byte(data))

	// Konversi byte menjadi string hexadecimal (64 karakter)
	return hex.EncodeToString(hash[:])
}
