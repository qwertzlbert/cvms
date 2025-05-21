package types

import (
	"encoding/json"
	"fmt"
	"time"
)

type VoteExtension struct {
	Address         string
	BlockCommitFlag int64
	VoteExtension   []byte
	Signature       []byte
}

type BlockSummary struct {
	BlockHeight           int64
	BlockTimeStamp        time.Time
	BlockProposerAddress  string
	Txs                   []Tx
	LastCommitBlockHeight int64
	BlockSignatures       []Signature
	CosmosValidators      []CosmosValidator
}

const (
	BOND_STATUS_BONDED      = "BOND_STATUS_BONDED"
	BOND_STATUS_UNBONDING   = "BOND_STATUS_UNBONDING"
	BOND_STATUS_UNBONDED    = "BOND_STATUS_UNBONDED"
	BOND_STATUS_UNSPECIFIED = "BOND_STATUS_UNSPECIFIED"
)

type BondStatus string

var (
	Bonded     BondStatus = BOND_STATUS_BONDED
	Unbonding  BondStatus = BOND_STATUS_UNBONDING
	Unbonded   BondStatus = BOND_STATUS_UNBONDED
	Unspecfied BondStatus = BOND_STATUS_UNSPECIFIED
)

// query path for cosmos status to check chain id
var CosmosStatusQueryPath = "/status"

// response type for v34 cosmos status
type CosmosV34StatusResponse struct {
	JsonRPC string       `json:"jsonrpc" validate:"required"`
	ID      int          `json:"id" validate:"required"`
	Result  CosmosStatus `json:"result" validate:"required"`
}

// response type for upper than v37 cosmos status
type CosmosV37StatusResponse CosmosStatus

// response of cosmos-sdk based chain status
type CosmosStatus struct {
	NodeInfo map[string]any `json:"node_info"`
	SyncInfo struct {
		LatestBlockHash     string    `json:"latest_block_hash"`
		LatestAppHash       string    `json:"latest_app_hash"`
		LatestBlockHeight   string    `json:"latest_block_height"`
		LatestBlockTime     time.Time `json:"latest_block_time"`
		EarliestBlockHash   string    `json:"earliest_block_hash"`
		EarliestAppHash     string    `json:"earliest_app_hash"`
		EarliestBlcokHeight string    `json:"earliest_block_height"`
		EarliestBlockTime   time.Time `json:"earliest_block_time"`
		CatchingUp          bool      `json:"catching_up"`
	} `json:"sync_info" validate:"required"`
	ValidatorInfo map[string]any `json:"validator_info"`
}

// query path for cosmos block by height
var CosmosBlockQueryPath = func(height int64) string {
	return fmt.Sprintf("/block?height=%d", height)
}

// response type for v34 cosmos block
type CosmosV34BlockResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  CosmosBlock `json:"result"`
}

// response type for v37 cosmos block
type CosmosV37BlockResponse CosmosBlock

// response of cosmos-sdk based chain block
type CosmosBlock struct {
	BlockID interface{} `json:"-"`
	Block   struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
		Data struct {
			Txs []Tx `json:"txs"`
		} `json:"data"`
		Evidence   interface{} `json:"-"`
		LastCommit struct {
			Height     string      `json:"height"`
			Round      uint64      `json:"-"`
			BlockID    interface{} `json:"-"`
			Signatures []Signature `json:"signatures"`
		} `json:"last_commit"`
	} `json:"block"`
}

type Tx string

// cosmos chain's validator signature type
type Signature struct {
	BlockIDFlag      int64       `json:"block_id_flag"`
	ValidatorAddress string      `json:"validator_address"`
	Timestamp        time.Time   `json:"timestamp"`
	Signature        interface{} `json:"signature"`
}

// query path for cosmos validator by height and page
var CosmosValidatorQueryPathWithHeight = func(height int64, page int) string {
	return fmt.Sprintf("/validators?height=%d&page=%d&per_page=100", height, page)
}

// query path for cosmos validator by height and page
var CosmosValidatorQueryPath = func(page int) string {
	return fmt.Sprintf("/validators?page=%d&per_page=100", page)
}

// ref; https://github.com/cosmos/cosmos-sdk/blob/v0.47.13/proto/cosmos/staking/v1beta1/staking.proto#L141
var CosmosStakingValidatorQueryPath = func(status string) string {
	return fmt.Sprintf("/cosmos/staking/v1beta1/validators?status=%s&pagination.count_total=true&pagination.limit=500", status)
}

// response type for v34 cosmos validators
type CosmosV34ValidatorResponse struct {
	Result CosmosValidators `json:"result"`
}

// response type for v37 cosmos validators
type CosmosV37ValidatorResponse CosmosValidators

// response of cosmos-sdk based chain validators
type CosmosValidators struct {
	BlockHeight string            `json:"block_height"`
	Validators  []CosmosValidator `json:"validators"`
	Total       string            `json:"total"`
}

// cosmos chain's validator type
type CosmosValidator struct {
	Address string `json:"address"`
	Pubkey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	VotingPower      string `json:"voting_power"`
	ProposerPriority string `json:"proposer_priority"`
}

// staking module
type StakingValidatorMetaInfo struct {
	Moniker         string
	OperatorAddress string
}

type CosmosStakingValidatorsQueryResponse struct {
	Validators []CosmosStakingValidator `json:"validators"`
	Pagination struct {
		// NextKey interface{} `json:"-"`
		Total string `json:"total"`
	} `json:"pagination"`
}

type CosmosStakingValidator struct {
	OperatorAddress string          `json:"operator_address"`
	ConsensusPubkey ConsensusPubkey `json:"consensus_pubkey"`
	Description     struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Commission struct {
		CommissionRates struct {
			Rate string `json:"rate"`
		} `json:"commission_rates"`
	} `json:"commission"`
	Tokens string `json:"tokens"`
	Status string `json:"status"`
}

type ConsensusPubkey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

// ccv module
// Returns the list of validators of a specific consumer chain
var (
	ProviderValidatorsQueryPath = func(consumerID string) string {
		return fmt.Sprintf("/interchain_security/ccv/provider/consumer_validators/%s", consumerID)
	}
)

type CosmosProviderValidatorsResponse struct {
	Validators []ProviderValidator `json:"validators"`
}
type ProviderValidator struct {
	PrvodierValconsAddress string `json:"provider_address"`
	ConsumerKey            struct {
		Pubkey string `json:"ed25519"`
	} `json:"consumer_key"`
	Description struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	ProviderValoperAddress string `json:"provider_operator_address"`
	Jailed                 bool   `json:"jailed"`
}

// see https://buf.build/cosmos/interchain-security/docs/main:interchain_security.ccv.provider.v1#interchain_security.ccv.provider.v1.QueryConsumerChainsRequest
// The phase of the consumer chains returned (optional)
// Registered=1|Initialized=2|Launched=3|Stopped=4|Deleted=5
var ConsumerChainListQueryPath string = "/interchain_security/ccv/provider/consumer_chains/3"

type CosmosConsumerChainsResponse struct {
	Chains []ConsumerChain `json:"chains"`
}

type ConsumerChain struct {
	ChainID    string `json:"chain_id"`
	ClientID   string `json:"client_id"`
	ConsumerID string `json:"consumer_id"`
}

// slashing module
var (
	CosmosSlashingLimitQueryPath  string = "/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1"
	CosmosSlashingParamsQueryPath string = "/cosmos/slashing/v1beta1/params"
	CosmosSlashingQueryPath              = func(consensusAddress string) string {
		return fmt.Sprintf("/cosmos/slashing/v1beta1/signing_infos/%s", consensusAddress)
	}
)

type CosmosSlashingResponse struct {
	ValidatorSigningInfo SigningInfo   `json:"val_signing_info"`
	Info                 []SigningInfo `json:"info"`
}

type SigningInfo struct {
	ConsensusAddress    string `json:"address"`
	StartHeight         string `json:"start_height"`
	IndexOffset         string `json:"index_offset"`
	JailedUntil         string `json:"jailed_until"`
	Tombstoned          bool   `json:"tombstoned"`
	MissedBlocksCounter string `json:"missed_blocks_counter"`
}

type CosmosSlashingParamsResponse struct {
	Params SlashingParam `json:"params"`
}

type SlashingParam struct {
	SignedBlocksWindow      string `json:"signed_blocks_window"`
	MinSignedPerWindow      string `json:"min_signed_per_window"`
	DowntimeJailDuration    string `json:"downtime_jail_duration"`
	SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
	SlashFractionDowntime   string `json:"slash_fraction_downtime"`
}

const CosmosUpgradeQueryPath = "/cosmos/upgrade/v1beta1/current_plan"

// ref; https://github.com/cosmos/cosmos-sdk/blob/main/x/upgrade/proto/cosmos/upgrade/v1beta1/query.proto#L14
type CosmosUpgradeResponse struct {
	Plan struct {
		Name                string `json:"name"`
		Time                string `json:"time"`
		Height              string `json:"height"`
		Info                string `json:"info"`
		UpgradedClientState string `json:"upgraded_client_state"`
	} `json:"plan"`
}

var CosmosBlockResultsQueryPath = func(height int64) string {
	return fmt.Sprintf("/block_results?height=%d", height)
}

type CosmosBlockResultResponse struct {
	JsonRPC string `json:"jsonrpc" validate:"required"`
	ID      int    `json:"id" validate:"required"`
	Result  struct {
		Height     string     `json:"height"`
		TxsResults []TxResult `json:"txs_results"`
		// case1)
		// https://github.com/cometbft/cometbft/blob/v0.37.0/rpc/core/types/responses.go#L54
		BeginBlockEvents []BlockEvent `json:"begin_block_events"`
		EndBlockEvents   []BlockEvent `json:"end_block_events"`
		// case2)
		// https://github.com/cometbft/cometbft/blob/v0.38.0/rpc/core/types/responses.go#L54
		FinalizeBlockEvents []BlockEvent `json:"finalize_block_events"`
		//
		ValidatorUpdate       interface{}           `json:"-"`
		ConsensusParamUpdates ConsensusParamUpdates `json:"consensus_param_updates"`
	} `json:"result" validate:"required"`
}

type ConsensusParamUpdates struct {
	Block struct {
		MaxBytes string `json:"max_bytes"`
		MaxGas   string `json:"max_gas"`
	}
}

type TxResult struct {
	Code      int64        `json:"code"`
	Data      string       `json:"data"`
	GasWanted string       `json:"gas_wanted"`
	GasUsed   string       `json:"gas_used"`
	Events    []BlockEvent `json:"events"`
}

type BlockEvent struct {
	TypeName   string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}

var (
	CosmosLimitedBlockTxsQueryPath = func(blockHeight int64) string {
		return fmt.Sprintf("/cosmos/tx/v1beta1/txs/block/%d?pagination.limit=1", blockHeight)
	}

	// This query suffers from a bug in the cosmos-sdk
	// that prevents the ‘next_key’ from being checked.
	// Therefore, it tries to lookup the entire tx at once using pagination.limit=1000.
	CosmosBlockTxsQueryPath = func(blockHeight int64) string {
		return fmt.Sprintf("/cosmos/tx/v1beta1/txs/block/%d?pagination.limit=1000", blockHeight)
	}
)

type CosmosBlockAndTxsResponse struct {
	Txs   []CosmosTx `json:"txs"`
	Block struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
	} `json:"block"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

type CosmosTx struct {
	Body struct {
		Messages []json.RawMessage `json:"messages"`
	} `json:"body"`
	AuthInfo   interface{} `json:"-"`
	Signatures []string    `json:"-"`
}

type CosmosErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details"`
}

type CosmosBlockData struct {
	TxResults []TxResult
	ConsensusParamUpdates
}

var CosmosConsensusParamsQueryPath = "/consensus_params"

type CosmosConsensusParams struct {
	Result struct {
		ConsensusParams struct {
			Block BlockParams `json:"block"`
			// Evidence  EvidenceParams  `json:"evidence"`
			// Validator ValidatorParams `json:"validator"`
			// Version   VersionParams   `json:"version"`
			// ABCI      ABCIParams      `json:"abci"`
		} `json:"consensus_params"`
	} `json:"result"`
}

type BlockParams struct {
	MaxBytes string `json:"max_bytes"`
	MaxGas   string `json:"max_gas"`
}
