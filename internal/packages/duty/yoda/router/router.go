package router

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/api"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/parser"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
)

func GetStatus(client *common.Exporter, chainName string) (types.CommonYodaStatus, error) {
	var (
		commonYodaQueryPath           string
		commonYodaParser              func(resp []byte) (isActive float64, err error)
		commonYodaParamsPath          string
		commonYodaParamsParser        func(resp []byte) (slashingWindow float64, err error)
		commonYodaRequestCountsPath   string
		commonYodaRequestCountsParser func(resp []byte) (requestCount float64, err error)
		commonYodaRequestPath         string
		commonYodaRequestParser       func(resp []byte) (requestBlock int64, validatorsFailedToRespond []string, status string, err error)
		commonYodaLatestBlockPath     string
		commonYodaLatestBlockParser   func(resp []byte) (latestBlock int64, err error)
	)

	switch chainName {
	case "band":
		commonYodaQueryPath = types.BandYodaQueryPath
		commonYodaParser = parser.BandYodaParser
		commonYodaParamsPath = types.BandYodaParamsPath
		commonYodaParamsParser = parser.BandYodaParamsParser
		commonYodaRequestCountsPath = types.BandYodaRequestCountsPath
		commonYodaRequestCountsParser = parser.BandYodaRequestCountParser
		commonYodaRequestPath = types.BandYodaRequestsPath
		commonYodaRequestParser = parser.BandYodaRequestParser
		commonYodaLatestBlockPath = types.BandLatestBlockHeightRequestPath
		commonYodaLatestBlockParser = parser.BandLatestBlockParser

		lastBlock, _ := api.GetBlockHeight(
			client,
			commonYodaLatestBlockPath,
			commonYodaLatestBlockParser)

		return api.GetYodaStatus(
			client,
			commonYodaQueryPath,
			commonYodaParser,
			commonYodaParamsPath,
			commonYodaParamsParser,
			commonYodaRequestCountsPath,
			commonYodaRequestCountsParser,
			commonYodaRequestPath,
			commonYodaRequestParser,
			lastBlock)
	default:
		return types.CommonYodaStatus{}, common.ErrOutOfSwitchCases
	}
}
