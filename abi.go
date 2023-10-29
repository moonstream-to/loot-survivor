package main

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
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

func HashFromName(name string) string {
	x := big.NewInt(0)
	mask := big.NewInt(0)

	x.Exp(big.NewInt(2), big.NewInt(250), nil)
	mask.Sub(x, big.NewInt(1))

	components := strings.Split(name, "::")
	eventName := components[len(components)-1]

	hashedName := crypto.Keccak256([]byte(eventName))
	hashedEncodedName := big.NewInt(0).SetBytes(hashedName)

	starknetHashedEncodedName := big.NewInt(0).And(hashedEncodedName, mask)
	return hex.EncodeToString(starknetHashedEncodedName.Bytes())
}

func Events(abi []map[string]interface{}) []SurvivorEvent {
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
			events[currentIndex] = SurvivorEvent{
				Name: item["name"].(string),
				Hash: HashFromName(item["name"].(string)),
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

	return events
}
