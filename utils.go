package main

import (
	"encoding/hex"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func AddressFelt(address string) (*felt.Felt, error) {
	fieldAdditiveIdentity := fp.NewElement(0)

	if address[:2] == "0x" {
		address = address[2:]
	}
	decodedAddress, decodeErr := hex.DecodeString(address)
	if decodeErr != nil {
		return nil, decodeErr
	}
	addressFelt := felt.NewFelt(&fieldAdditiveIdentity)
	addressFelt.SetBytes(decodedAddress)

	return addressFelt, nil
}
