package builtin

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// Hash provides methods to compute hash values using various algorithms.
type Hash struct{}

// NewHash creates a new Hash instance.
// NewHash is mapped to the built-in function `hash`.
func NewHash() *Hash {
	return &Hash{}
}

// Sha256 computes SHA-256 hash of the input data and returns it as a hex string.
func (h *Hash) Sha256(v any) (string, error) {
	data := h.toBytes(v)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Sha512 computes SHA-512 hash of the input data and returns it as a hex string.
func (h *Hash) Sha512(v any) (string, error) {
	data := h.toBytes(v)
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:]), nil
}

// toBytes converts input value to byte slice.
func (h *Hash) toBytes(v any) []byte {
	switch vv := v.(type) {
	case string:
		return []byte(vv)
	case []byte:
		return vv
	default:
		return []byte(fmt.Sprintf("%v", vv))
	}
}
