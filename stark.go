package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type RPCRequest struct {
	RPCVersion string        `json:"jsonrpc"`
	ID         uint64        `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params,omitempty"`
}

type Starknet struct {
	ProviderURL string
	Client      *http.Client
	RPCVersion  string
	// Note: This doesn't properly handle header keys with multiple values
	Headers          map[string]string
	CurrentRequestID uint64
}

func (provider *Starknet) RPCMethod(method string, params []interface{}) (interface{}, error) {
	requestRaw := RPCRequest{
		RPCVersion: provider.RPCVersion,
		ID:         provider.CurrentRequestID,
		Method:     method,
		Params:     params,
	}

	fmt.Printf("Raw request: %v\n", requestRaw)

	buffer := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(buffer)

	encoder.Encode(requestRaw)

	request, makeRequestErr := http.NewRequest("POST", provider.ProviderURL, buffer)
	if makeRequestErr != nil {
		return nil, makeRequestErr
	}

	for k, v := range provider.Headers {
		request.Header.Add(k, v)
	}

	response, responseErr := provider.Client.Do(request)
	if responseErr != nil {
		return nil, responseErr
	}
	defer response.Body.Close()

	var result interface{}
	decodeErr := json.NewDecoder(response.Body).Decode(&result)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return result, nil
}

// Returns the current block number on the given provider.
func (provider *Starknet) Blocknumber() (uint64, error) {
	result, err := provider.RPCMethod("starknet_blockNumber", []interface{}{})
	if err != nil {
		return 0, err
	}

	return uint64(result.(map[string]interface{})["result"].(float64)), nil
}
