package main

import (
	"encoding/hex"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func FeltFromHexString(hexString string) (*felt.Felt, error) {
	fieldAdditiveIdentity := fp.NewElement(0)

	if hexString[:2] == "0x" {
		hexString = hexString[2:]
	}
	decodedString, decodeErr := hex.DecodeString(hexString)
	if decodeErr != nil {
		return nil, decodeErr
	}
	derivedFelt := felt.NewFelt(&fieldAdditiveIdentity)
	derivedFelt.SetBytes(decodedString)

	return derivedFelt, nil
}
