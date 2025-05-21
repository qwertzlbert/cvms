package api

import (
	"context"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/pkg/errors"
)

func GetCosmosConsensusParams(c common.CommonClient) (float64, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.RPCClient
	resp, err := requester.Get(ctx, types.CosmosConsensusParamsQueryPath)
	if err != nil {
		endpoint, _ := requester.GetEndpoint()
		return 0, 0, errors.Errorf("rpc call is failed from %s: %s", endpoint, err)
	}

	maxBytes, maxGas, err := parser.CosmosConsensusmParamsParser(resp)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return maxBytes, maxGas, nil
}
