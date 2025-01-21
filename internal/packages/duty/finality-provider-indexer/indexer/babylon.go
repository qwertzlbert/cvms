package indexer

import (
	"encoding/json"
	"fmt"
)

// Define the structures for the JSON data
type SigningInfo struct {
	FPBTCPkHex          string `json:"fp_btc_pk_hex"`
	StartHeight         string `json:"start_height"`
	MissedBlocksCounter string `json:"missed_blocks_counter"`
	JailedUntil         string `json:"jailed_until"`
}

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

type SigningInfoResponse struct {
	SigningInfos []SigningInfo `json:"signing_infos"`
	Pagination   Pagination    `json:"pagination"`
}

func ParserFinalityProviderSigningInfos(resp []byte) (SigningInfoResponse, error) {
	var result SigningInfoResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return SigningInfoResponse{}, nil
	}

	return result, nil
}

var BabylonFinalityProvidersQueryPath = func(height int64) string {
	return fmt.Sprintf("/babylon/finality/v1/finality_providers/%d", height)
}

type FinalityProvider struct {
	BtcPkHex             string `json:"btc_pk_hex"`
	Height               string `json:"height"`
	VotingPower          string `json:"voting_power"`
	SlashedBabylonHeight string `json:"slashed_babylon_height"`
	SlashedBtcHeight     int    `json:"slashed_btc_height"`
	Jailed               bool   `json:"jailed"`
	HighestVotedHeight   int    `json:"highest_voted_height"`
}

type FinalityProvidersResponse struct {
	FinalityProviders []FinalityProvider `json:"finality_providers"`
	Pagination        Pagination         `json:"pagination"`
}

func ParseFinalityProviders(resp []byte) (FinalityProvidersResponse, error) {
	var result FinalityProvidersResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return FinalityProvidersResponse{}, nil
	}

	return result, nil
}

var BabylonFinalityVotesQueryPath = func(height int64) string {
	return fmt.Sprintf("/babylon/finality/v1/votes/%d", height)
}

type FinalityVotesResponse struct {
	BTCPKs []string `json:"btc_pks"`
}

func ParseFinalityProviderVotings(resp []byte) (FinalityVotesResponse, error) {
	var result FinalityVotesResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return FinalityVotesResponse{}, nil
	}

	return result, nil
}

var BabylonFinalityProviderInfosQueryPath = func(key string) string {
	return fmt.Sprintf("/babylon/btcstaking/v1/finality_providers?pagination.key=%s", key)
}

type FinalityProviderInfosResponse struct {
	FinalityProviders []FinalityProviderInfo `json:"finality_providers"`
	Pagination        Pagination             `json:"pagination"`
}

type FinalityProviderInfo struct {
	Description struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Address string `json:"addr"`
	BTCPK   string `json:"btc_pk"`
}

func ParseFinalityProviderInfos(resp []byte) (FinalityProviderInfosResponse, error) {
	var result FinalityProviderInfosResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return FinalityProviderInfosResponse{}, nil
	}

	return result, nil
}

type fpVoteMap map[string]int64

type FinalityVoteSummary struct {
	BlockHeight           int64
	FinalityProviderVotes fpVoteMap
}
