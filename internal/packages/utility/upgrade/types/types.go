package types

var (
	SupportedProtocolTypes = []string{"cosmos"}
)

type CommonUpgrade struct {
	RemainingTime   float64
	RemainingBlocks float64
	UpgradeName     string
}
