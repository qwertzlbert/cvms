package router

import (
	"github.com/cosmostation/cvms/internal/common"

	commonapi "github.com/cosmostation/cvms/internal/common/api"
	commonparser "github.com/cosmostation/cvms/internal/common/parser"
	commontypes "github.com/cosmostation/cvms/internal/common/types"

	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/api"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/types"
)

func GetStatus(exporter *common.Exporter, p common.Packager) (types.CommonUptimeStatus, error) {
	switch p.ProtocolType {
	case "cosmos":
		var (
			commonSlashingValidatorQueryPath   string
			commonSlashingValidatorQueryParser func(resp []byte) (
				consensusAddress string,
				indexOffset float64,
				isTomstoned float64,
				issedBlocksCounter float64,
				err error)
		)

		// if p.IsConsumerChain {
		// 	return api.GetConsumserUptimeStatus(exporter, p.ChainID)
		// }

		commonSlashingValidatorQueryPath = commontypes.CosmosSlashingQueryPath
		commonSlashingValidatorQueryParser = commonparser.CosmosSlashingParser

		stakingValidators, _ := commonapi.GetStakingValidators(exporter.CommonClient, p.ChainName)
		consensusValidators, _ := commonapi.GetValidators(exporter.CommonClient)
		validatorUptimeStatus, _ := api.GetValidatorUptimeStatus(
			exporter.CommonApp,
			commonSlashingValidatorQueryPath,
			commonSlashingValidatorQueryParser,
			consensusValidators,
			stakingValidators)
		signedBlocksWindow, minSignedPerWindow, _ := api.GetUptimeParams(exporter.CommonClient, p.ChainName)
		return api.GetUptimeStatus(signedBlocksWindow, minSignedPerWindow, validatorUptimeStatus)
	default:
		return types.CommonUptimeStatus{}, common.ErrOutOfSwitchCases
	}
}
