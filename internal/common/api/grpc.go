package api

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/types"
)

func GetValidatorsbyGRPC(c common.CommonClient, height ...chan int64) ([]types.CosmosValidator, error) {
	return nil, nil // Placeholder return value
}
