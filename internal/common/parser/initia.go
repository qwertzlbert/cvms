package parser

import (
	"encoding/json"
	"strconv"

	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/pkg/errors"
)

func InitiaStakingValidatorParser(resp []byte) ([]types.CosmosStakingValidator, int64, error) {
	var result types.InitiaStakingValidatorsQueryResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, 0, errors.Cause(err)
	}
	commonStakingValidators := make([]types.CosmosStakingValidator, 0)
	for _, validator := range result.Validators {
		commonStakingValidators = append(commonStakingValidators, types.CosmosStakingValidator{
			OperatorAddress: validator.OperatorAddress,
			ConsensusPubkey: validator.ConsensusPubkey,
			Description:     validator.Description,
			Tokens:          "", // initia has multiple tokens on validators, so skip the tokens
		})
	}

	valCount, err := strconv.ParseInt(result.Pagination.Total, 10, 64)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to convert from stirng to float in parser")
	}

	return commonStakingValidators, valCount, nil
}
