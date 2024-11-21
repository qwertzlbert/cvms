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
	TotalMisses              float64 `json:"total_misses"`
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
