package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign produces an HMAC-SHA256 signature for payload using secret.
// All A2AMessages are signed before publishing and verified by PlannerAgent.
// This prevents rogue agent injection (spec requirement).
func Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify returns true if sig matches HMAC-SHA256(payload, secret).
func Verify(payload []byte, secret string, sig string) bool {
	expected := Sign(payload, secret)
	return hmac.Equal([]byte(expected), []byte(sig))
}
