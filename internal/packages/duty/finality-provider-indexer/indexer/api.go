package indexer

import (
	"context"

	"github.com/cosmostation/cvms/internal/common"
	indexmodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	commonparser "github.com/cosmostation/cvms/internal/common/parser"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/pkg/errors"
)

func GetFinalityProvidersInfo(c common.CommonClient, newFinalityProviderMap map[string]bool, chainInfoID int64) ([]indexmodel.FinalityProviderInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	fpInfoList := make([]indexmodel.FinalityProviderInfo, 0)
	maxCnt := 10
	key := ""
	for cnt := 0; cnt <= maxCnt; cnt++ {
		resp, err := c.APIClient.Get(ctx, commontypes.BabylonFinalityProviderInfosQueryPath(key))
		if err != nil {
			return nil, errors.Errorf("rpc call is failed from %s: %s", commontypes.BabylonFinalityProviderInfosQueryPath(key), err)
		}

		fpInfos, err := commonparser.ParseFinalityProviderInfos(resp)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, fp := range fpInfos.FinalityProviders {
			fpInfoList = append(fpInfoList, indexmodel.FinalityProviderInfo{
				ChainInfoID:     chainInfoID,
				Moniker:         fp.Description.Moniker,
				BTCPKs:          fp.BTCPK,
				OperatorAddress: fp.Address,
			})
		}

		if fpInfos.Pagination.NextKey != "" {
			key = fpInfos.Pagination.NextKey
		} else {
			// got all finality provider infos
			break
		}
	}

	return fpInfoList, nil
}
