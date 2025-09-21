package hashutils

import "encoding/hex"

func HashPayload(b []byte) string {
	hashed := hex.EncodeToString(b)
	return hashed
}
