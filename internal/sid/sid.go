// Package sid generates stable unique identifiers for linemark entities.
package sid

import (
	"fmt"
	"io"
)

const (
	alphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	alphabetSize = len(alphabet) // 62
	sidLength    = 12
	threshold    = alphabetSize * (256 / alphabetSize) // largest multiple of alphabetSize ≤ 256
)

// Generate produces a 12-character base62 SID by reading random bytes from r.
// It uses rejection sampling to ensure unbiased distribution: bytes >= 248
// are discarded and re-read.
func Generate(r io.Reader) (string, error) {
	var buf [1]byte
	result := make([]byte, 0, sidLength)

	for len(result) < sidLength {
		_, err := r.Read(buf[:])
		if err != nil {
			return "", fmt.Errorf("reading random byte: %w", err)
		}
		b := int(buf[0])
		if b >= threshold {
			continue
		}
		result = append(result, alphabet[b%alphabetSize])
	}

	return string(result), nil
}
