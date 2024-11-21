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
		CommonYodaRequestCountsPath   string
		CommonYodaRequestCountsParser func(resp []byte) (requestCount float64, err error)
	)

	switch chainName {
	case "band":
		commonYodaQueryPath = types.BandYodaQueryPath
		commonYodaParser = parser.BandYodaParser
		commonYodaParamsPath = types.BandYodaParamsPath
		commonYodaParamsParser = parser.BandYodaParamsParser
		CommonYodaRequestCountsPath = types.BandYodaRequestCountsPath
		CommonYodaRequestCountsParser = parser.BandYodaRequestCountParser

		return api.GetYodaStatus(client, commonYodaQueryPath, commonYodaParser, commonYodaParamsPath, commonYodaParamsParser, CommonYodaRequestCountsPath, CommonYodaRequestCountsParser)
	default:
		return types.CommonYodaStatus{}, common.ErrOutOfSwitchCases
	}
}
