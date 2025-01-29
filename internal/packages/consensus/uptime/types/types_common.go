package types

var (
	SupportedValconsTypes  = []string{"valcons", "ica"}
	SupportedProtocolTypes = []string{"cosmos"}
)

// common
type CommonUptimeStatus struct {
	MinSignedPerWindow      float64                 `json:"slash_winodw"`
	SignedBlocksWindow      float64                 `json:"vote_period"`
	DowntimeJailDuration    float64                 `json:"downtime_jail_duration"`
	SlashFractionDowntime   float64                 `json:"slash_fraction_downtime"`
	SlashFractionDoubleSign float64                 `json:"slash_fraction_double_sign"`
	BondedValidatorsTotal   int                     `json:"bonded_validators_total"`
	ActiveValidatorsTotal   int                     `json:"active_validators_total"`
	MinimumSeatPrice        int64                   `json:"minimum_seat_price"`
	Validators              []ValidatorUptimeStatus `json:"validators"`
}

// cosmos uptime status
type ValidatorUptimeStatus struct {
	Moniker                   string  `json:"moniker"`
	ProposerAddress           string  `json:"proposer_address"`
	ValidatorOperatorAddress  string  `json:"validator_operator_address"`
	ValidatorConsensusAddress string  `json:"validator_consensus_addreess"`
	MissedBlockCounter        float64 `json:"missed_block_counter"`
	VotingPower               float64
	IsTomstoned               float64
	StakedTokens              int
	CommissionRate            float64
	// Only Consumer Chain
	ConsumerConsensusAddress string `json:"consumer_consensus_address"`
}
