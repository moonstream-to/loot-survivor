package main

import (
	"encoding/hex"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

type SurvivorEventData struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Kind string `json:"kind"`
}

type SurvivorEvent struct {
	Name string              `json:"name"`
	Hash string              `json:"hash"`
	Data []SurvivorEventData `json:"data"`
}

func HashFromName(name string) (string, error) {
	x := big.NewInt(0)
	mask := big.NewInt(0)

	x.Exp(big.NewInt(2), big.NewInt(250), nil)
	mask.Sub(x, big.NewInt(1))

	components := strings.Split(name, "::")
	eventName := components[len(components)-1]

	// Very important to use the LegacyKeccak256 here - to match Ethereum:
	// https://pkg.go.dev/golang.org/x/crypto/sha3#NewLegacyKeccak256
	hash := sha3.NewLegacyKeccak256()
	_, hashErr := hash.Write([]byte(eventName))
	if hashErr != nil {
		return "", hashErr
	}

	b := make([]byte, 0)
	hashedNameBytes := hash.Sum(b)

	hashedEncodedName := big.NewInt(0).SetBytes(hashedNameBytes)

	starknetHashedEncodedName := big.NewInt(0).And(hashedEncodedName, mask)
	return hex.EncodeToString(starknetHashedEncodedName.Bytes()), nil
}

func Events(abi []map[string]interface{}) ([]SurvivorEvent, error) {
	numEvents := 0
	for _, item := range abi {
		if item["type"] == "event" && item["kind"] == "struct" {
			numEvents++
		}
	}

	currentIndex := 0
	events := make([]SurvivorEvent, numEvents)
	for _, item := range abi {
		if item["type"] == "event" && item["kind"] == "struct" {
			name := item["name"].(string)
			hashedName, hashErr := HashFromName(name)
			if hashErr != nil {
				return nil, hashErr
			}
			events[currentIndex] = SurvivorEvent{
				Name: name,
				Hash: hashedName,
			}
			if item["members"] != nil {
				membersArray := item["members"].([]interface{})
				membersMapArray := make([]map[string]interface{}, len(membersArray))
				for i, member := range membersArray {
					membersMapArray[i] = member.(map[string]interface{})
				}
				events[currentIndex].Data = make([]SurvivorEventData, len(membersArray))
				for i, member := range membersMapArray {
					events[currentIndex].Data[i] = SurvivorEventData{
						Name: member["name"].(string),
						Type: member["type"].(string),
						Kind: member["kind"].(string),
					}
				}
			}

			currentIndex++
		}
	}

	return events, nil
}
