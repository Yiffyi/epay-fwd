package sec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
)

func DeriveMyEpayKey(pid int, fwdSecret string) string {
	// Convert integer to bytes
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf, uint32(pid))

	// Create HMAC using the secret as key
	h := hmac.New(sha256.New, []byte(fwdSecret))
	h.Write(buf)
	hash := h.Sum(nil)

	// Return as URL-safe base64
	return base64.RawURLEncoding.EncodeToString(hash)
}
