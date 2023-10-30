package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var ErrAddressIsNotContract error = errors.New("address is not a contract")
var ErrPotentialReorg error = errors.New("potential reorg")

// Perform a binary search to determine the block number at which the contract at the given address
// was deployed.
// Since the starknet_getCode method has been deprecated, this uses starknet_getClassHashAt in order
// to conduct the search. If the contract has not been deployed at a given block, calling
// starknet_getClassHashAt at that block will result in an error with code 20.
func DeploymentBlock(ctx context.Context, provider *rpc.Provider, address *felt.Felt) (uint64, error) {
	maxBlock, blockNumberErr := provider.BlockNumber(ctx)
	if blockNumberErr != nil {
		return 0, blockNumberErr
	}

	var minBlock uint64 = 0

	midBlock := (minBlock + maxBlock) / 2

	var isDeployed map[uint64]bool = make(map[uint64]bool)

	isDeployedAtBlock, blockErr := ContractExistsAtBlock(ctx, provider, address, maxBlock)
	if blockErr != nil {
		return 0, blockErr
	}
	if !isDeployedAtBlock {
		return 0, ErrAddressIsNotContract
	}
	isDeployed[maxBlock] = isDeployedAtBlock

	isDeployed[minBlock], blockErr = ContractExistsAtBlock(ctx, provider, address, minBlock)
	if blockErr != nil {
		return 0, blockErr
	}

	isDeployed[midBlock], blockErr = ContractExistsAtBlock(ctx, provider, address, midBlock)
	if blockErr != nil {
		return 0, blockErr
	}

	for (maxBlock - minBlock) >= 2 {
		if !isDeployed[minBlock] && !isDeployed[midBlock] {
			minBlock = midBlock
		} else {
			maxBlock = midBlock
		}

		midBlock = (minBlock + maxBlock) / 2

		isDeployed[midBlock], blockErr = ContractExistsAtBlock(ctx, provider, address, midBlock)
		if blockErr != nil {
			return 0, blockErr
		}
	}

	if isDeployed[minBlock] {
		return minBlock, nil
	}
	return maxBlock, nil
}

func ContractExistsAtBlock(ctx context.Context, provider *rpc.Provider, address *felt.Felt, blockNumber uint64) (bool, error) {
	_, err := provider.ClassHashAt(ctx, rpc.BlockID{Number: &blockNumber}, address)
	if err != nil {
		// Note: No other comparison (e.g. using errors.Is) is working.
		if err.Error() == rpc.ErrContractNotFound.Error() {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func AllEventsFilter(fromBlock, toBlock uint64, contractAddress string) (*rpc.EventFilter, error) {
	result := rpc.EventFilter{FromBlock: rpc.BlockID{Number: &fromBlock}, ToBlock: rpc.BlockID{Number: &toBlock}}

	fieldAdditiveIdentity := fp.NewElement(0)

	if contractAddress != "" {
		if contractAddress[:2] == "0x" {
			contractAddress = contractAddress[2:]
		}
		decodedAddress, decodeErr := hex.DecodeString(contractAddress)
		if decodeErr != nil {
			return &result, decodeErr
		}
		result.Address = felt.NewFelt(&fieldAdditiveIdentity)
		result.Address.SetBytes(decodedAddress)
	}

	result.Keys = [][]*felt.Felt{{}}

	return &result, nil
}

func SingleEventFilter(fromBlock, toBlock uint64, contractAddress, eventName string, abi []map[string]interface{}) (*rpc.EventFilter, error) {
	result, initialFilterErr := AllEventsFilter(fromBlock, toBlock, contractAddress)
	if initialFilterErr != nil {
		return result, initialFilterErr
	}

	abiEvents, abiErr := Events(abi)
	if abiErr != nil {
		return result, abiErr
	}

	eventHash := ""
	for _, event := range abiEvents {
		if event.Name == eventName {
			eventHash = event.Hash
			break
		}
	}
	if eventHash == "" {
		return result, ErrNoSuchEventInABI
	}

	decodedEventHash, decodeErr := hex.DecodeString(eventHash)
	if decodeErr != nil {
		return result, decodeErr
	}
	fieldAdditiveIdentity := fp.NewElement(0)
	eventKeyFelt := felt.NewFelt(&fieldAdditiveIdentity)
	eventKeyFelt.SetBytes(decodedEventHash)
	result.Keys = [][]*felt.Felt{
		{eventKeyFelt},
	}

	return result, nil
}

type CrawledEvent struct {
	BlockNumber     uint64
	BlockHash       *felt.Felt
	TransactionHash *felt.Felt
	FromAddress     *felt.Felt
	Parameters      []*felt.Felt
}

func ContractEvents(ctx context.Context, provider *rpc.Provider, contractAddress string, outChan chan<- CrawledEvent, hotThreshold int, hotInterval, coldInterval time.Duration, fromBlock, toBlock uint64, confirmations, batchSize int) error {
	defer func() { close(outChan) }()

	type CrawlCursor struct {
		FromBlock         uint64
		ToBlock           uint64
		ContinuationToken string
		Interval          time.Duration
		Heat              int
	}

	cursor := CrawlCursor{FromBlock: fromBlock, ToBlock: toBlock, ContinuationToken: "", Interval: hotInterval, Heat: 0}

	count := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(cursor.Interval):
			count++
			if cursor.ToBlock == 0 {
				currentblock, blockErr := provider.BlockNumber(ctx)
				if blockErr != nil {
					return blockErr
				}
				cursor.ToBlock = currentblock - uint64(confirmations)
			}

			if cursor.ToBlock <= cursor.FromBlock {
				// Crawl is cold, slow things down.
				cursor.Interval = coldInterval

				if toBlock == 0 {
					// If the crawl is continuous, breaks out of select, not for loop.
					// This effects a wait for the given interval.
					break
				} else {
					// If crawl is not continuous, just ends the crawl.
					return nil
				}
			}

			filter, filterErr := AllEventsFilter(cursor.FromBlock, cursor.ToBlock, contractAddress)
			if filterErr != nil {
				return filterErr
			}

			eventsInput := rpc.EventsInput{
				EventFilter:       *filter,
				ResultPageRequest: rpc.ResultPageRequest{ChunkSize: batchSize, ContinuationToken: cursor.ContinuationToken},
			}

			eventsChunk, getEventsErr := provider.Events(ctx, eventsInput)
			if getEventsErr != nil {
				return getEventsErr
			}

			for _, event := range eventsChunk.Events {
				crawledEvent := CrawledEvent{
					BlockNumber:     event.BlockNumber,
					BlockHash:       event.BlockHash,
					TransactionHash: event.TransactionHash,
					FromAddress:     event.FromAddress,
					Parameters:      event.Data,
				}

				outChan <- crawledEvent
			}

			if eventsChunk.ContinuationToken != "" {
				cursor.ContinuationToken = eventsChunk.ContinuationToken
				cursor.Interval = hotInterval
			} else {
				fmt.Fprintf(os.Stderr, "From: %d, To: %d\n", cursor.FromBlock, cursor.ToBlock)
				cursor.FromBlock = cursor.ToBlock + 1
				cursor.ToBlock = toBlock
				cursor.ContinuationToken = ""
				fmt.Fprintf(os.Stderr, "From: %d, To: %d\n", cursor.FromBlock, cursor.ToBlock)
				if len(eventsChunk.Events) > 0 {
					cursor.Heat++
					if cursor.Heat >= hotThreshold {
						cursor.Interval = hotInterval
					}
				} else {
					cursor.Heat = 0
					cursor.Interval = coldInterval
				}
			}
		}
	}
}
