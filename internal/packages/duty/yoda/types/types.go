package types

import "time"

var (
	SupportedChains = []string{"band"}
)

const (
	// common
	CommonValidatorQueryPath = "/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.count_total=true&pagination.limit=500"

	// band paths
	BandYodaQueryPath = "/oracle/v1/validators/{validator_address}"

	// band oracle params paths
	BandYodaParamsPath = "/oracle/v1/params"

	// band oracle path to get total request counts
	BandYodaRequestCountsPath = "/oracle/v1/counts"

	// band oracle path to get request details by request ID
	BandYodaRequestsPath = "/oracle/v1/requests/{request_id}"

	// band latest blockheight request
	BandLatestBlockHeightRequestPath = "/cosmos/base/tendermint/v1beta1/blocks/latest"
)

// common
type CommonYodaStatus struct {
	SlashWindow  float64 `json:"slash_window"`
	RequestCount float64 `json:"request_count"`
	Validators   []ValidatorStatus
}

type ValidatorStatus struct {
	Moniker                  string  `json:"moniker"`
	ValidatorOperatorAddress string  `json:"validator_operator_address"`
	IsActive                 float64 `json:"is_active"`
	MaxMisses                float64 `json:"max_misses"`
	Requests                 []RequestStatus
}

type CommonValidatorsQueryResponse struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		Description     struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
	} `json:"validators"`
	Pagination struct {
		NextKey interface{} `json:"-"`
		Total   string      `json:"-"`
	} `json:"-"`
}

// band
type BandYodaResponse struct {
	Status struct {
		IsActive bool      `json:"is_active"`
		Since    time.Time `json:"since"`
	} `json:"status"`
}

type BandYodaParamsResponse struct {
	Params struct {
		SlashWindow string `json:"expiration_block_count"`
	} `json:"params"`
}

type BandYodaRequestCountResponse struct {
	RequestCount string `json:"request_count"`
}

type BandLatestBlockHeightResponse struct {
	Block struct {
		Header struct {
			BlockHeight string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

type BandYodaRequestResponse struct {
	Request *struct {
		RequestBlock        string   `json:"request_height"`
		RequestedValidators []string `json:"requested_validators"`
	} `json:"request"`
	Reports []struct {
		Validator string `json:"validator,omitempty"`
	} `json:"reports"`
	Result *struct {
		ResolveStatus string `json:"resolve_status"`
	} `json:"result"`
}

type RequestStatus struct {
	RequestID                 int64    `json:"request_id"`
	Status                    string   `json:"status"` // three possible values: "running", "completed", "expired"
	RequestHeight             int64    `json:"request_height"`
	BlocksPassed              int64    `json:"blocks_passed"`
	ValidatorsFailedToRespond []string `json:"validators_failed_to_respond"` // list of validator addresses
}

type MissedRequests struct {
	Validator ValidatorStatus
	Request   RequestStatus
}
