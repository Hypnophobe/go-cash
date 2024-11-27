package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

func generateAddress(pkey string) string {
	sum := sha256.Sum256([]byte(pkey))
	addressHex := hex.EncodeToString(sum[:])
	address := addressHex[:12]

	return address
}

func validateAddress(address string) bool {
	_, err := strconv.ParseUint(address, 16, 64)
	return err == nil
}

func genBlock(block string, address string, nonce string) string {
	sum := sha256.Sum256([]byte(block + address + nonce))
	addressHex := hex.EncodeToString(sum[:])

	return addressHex
}
