package utils

import (
	"crypto/rand"
	"encoding/binary"
	"math"
)

func CryptoFloat64() (float64, error) {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return float64(binary.BigEndian.Uint64(b[:])) / math.MaxUint64, nil
}
