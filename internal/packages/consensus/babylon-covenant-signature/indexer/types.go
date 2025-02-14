package indexer

import (
	"encoding/json"
	"fmt"
	"time"
)

// SUPPORTED_MESSAGE_TYPES in CVMS
const (
	BabylonCovenantSignatureMessageType         = "/babylon.btcstaking.v1.MsgAddCovenantSigs"
	BabylonCovenantSignatureReceivedMessageType = "babylon.btcstaking.v1.EventCovenantSignatureReceived"
)

var (
	// TODO: move into common api
	BlockTxsQueryPath = func(blockHeight int64) string {
		return fmt.Sprintf("/cosmos/tx/v1beta1/txs/block/%d?pagination.limit=1", blockHeight)
	}
)

type MsgCovenantSignature struct {
	Type                    string   `json:"@type"`
	Signer                  string   `json:"signer"`
	Pk                      string   `json:"pk"`
	StakingTxHash           string   `json:"staking_tx_hash"`
	SlashingTxSigs          []string `json:"slashing_tx_sigs"`
	UnbondingTxSig          string   `json:"unbonding_tx_sig"`
	SlashingUnbondingTxSigs []string `json:"slashing_unbonding_tx_sigs"`
}

type CovenantSignature struct {
	BlockHeight                int64
	CovenantPk                 string
	BTCStakingTxHash           string
	CovenantUnbondingSignature string
	Timestamp                  time.Time
}

// TODO: I think this types should move into common cosmos types
type CosmosTx struct {
	Body struct {
		Messages []json.RawMessage `json:"messages"`
	} `json:"body"`
	AuthInfo   interface{} `json:"-"`
	Signatures []string    `json:"-"`
}

type BlockTxsResponse struct {
	Txs   []CosmosTx `json:"txs"`
	Block struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
	} `json:"block"`
}
