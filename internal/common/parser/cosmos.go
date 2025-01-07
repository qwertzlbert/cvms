package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/types"
	sdkhelper "github.com/cosmostation/cvms/internal/helper/sdk"
	"github.com/pkg/errors"
)

func CosmosBlockParser(resp []byte) (
	/* block height */ int64,
	/* block timestamp */ time.Time,
	/* block proposer addrss */ string,
	/* txs in the block */ []types.Tx,
	/* last comit block height*/ int64,
	/* block validators signatures */ []types.Signature,
	error,
) {
	var preResult map[string]interface{}
	if err := json.Unmarshal(resp, &preResult); err != nil {
		return 0, time.Time{}, "", nil, 0, nil, err
	}

	_, ok := preResult["jsonrpc"].(string)
	if ok { // v0.34.x
		var resultV34 types.CosmosV34BlockResponse
		if err := json.Unmarshal(resp, &resultV34); err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		heightString, blockTimestamp, lastCommitHeightString := resultV34.Result.Block.Header.Height, resultV34.Result.Block.Header.Time, resultV34.Result.Block.LastCommit.Height

		blockHeight, err := strconv.ParseInt(heightString, 10, 64)
		if err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		lastCommitBlockHeight, err := strconv.ParseInt(lastCommitHeightString, 10, 64)
		if err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		txs := resultV34.Result.Block.Data.Txs
		signatures := resultV34.Result.Block.LastCommit.Signatures
		proposerAddress := resultV34.Result.Block.Header.ProposerAddress
		return blockHeight, blockTimestamp, proposerAddress, txs, lastCommitBlockHeight, signatures, nil
	} else { // tendermint v0.37.x
		var resultV37 types.CosmosV37BlockResponse
		if err := json.Unmarshal(resp, &resultV37); err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		heightString, blockTimestamp, lastCommitHeightString := resultV37.Block.Header.Height, resultV37.Block.Header.Time, resultV37.Block.LastCommit.Height

		blockHeight, err := strconv.ParseInt(heightString, 10, 64)
		if err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		lastCommitBlockHeight, err := strconv.ParseInt(lastCommitHeightString, 10, 64)
		if err != nil {
			return 0, time.Time{}, "", nil, 0, nil, err
		}

		txs := resultV37.Block.Data.Txs
		signatures := resultV37.Block.LastCommit.Signatures
		proposerAddress := resultV37.Block.Header.ProposerAddress
		return blockHeight, blockTimestamp, proposerAddress, txs, lastCommitBlockHeight, signatures, nil
	}
}

func CosmosStatusParser(resp []byte) (
	/* latest block height */ int64,
	/* latest block timestamp */ time.Time,
	/* unexpected error */ error,
) {
	var preResult map[string]interface{}
	if err := json.Unmarshal(resp, &preResult); err != nil {
		return 0, time.Time{}, errors.Wrap(err, "failed to unmarshal json in parser")
	}

	_, ok := preResult["jsonrpc"].(string)
	if ok {
		var result types.CosmosV34StatusResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			return 0, time.Time{}, errors.Wrap(err, "failed to unmarshal json in parser")
		}

		blockTimestamp := result.Result.SyncInfo.LatestBlockTime
		blockHeight, err := strconv.ParseInt(result.Result.SyncInfo.LatestBlockHeight, 10, 64)
		if err != nil {
			return 0, time.Time{}, errors.Wrap(err, "failed to convert from stirng to float in parser")
		}

		return blockHeight, blockTimestamp, nil
	} else {
		var result types.CosmosV37StatusResponse
		if err := json.Unmarshal(resp, &result); err != nil {
			return 0, time.Time{}, errors.Wrap(err, "failed to unmarshal json in parser")
		}

		blockTimestamp := result.SyncInfo.LatestBlockTime
		blockHeight, err := strconv.ParseInt(result.SyncInfo.LatestBlockHeight, 10, 64)
		if err != nil {
			return 0, time.Time{}, errors.Wrap(err, "failed to convert from stirng to float in parser")
		}

		return blockHeight, blockTimestamp, nil
	}
}

// TODO: modify this function logic
func CosmosValidatorParser(resp []byte) ([]types.CosmosValidator, int64, error) {
	var validators types.CosmosV34ValidatorResponse
	err := json.Unmarshal(resp, &validators)
	if err != nil {
		return []types.CosmosValidator{}, 0, err
	}
	if len(validators.Result.Validators) == 0 {
		var validators types.CosmosV37ValidatorResponse
		err := json.Unmarshal(resp, &validators)
		if err != nil {
			return []types.CosmosValidator{}, 0, err
		}

		var count string
		if validators.Total != "" {
			count = validators.Total
		} else {
			count = validators.Pagination.Total
		}

		totalValidatorsCount, err := strconv.ParseInt(count, 10, 64)
		if err != nil {
			return []types.CosmosValidator{}, 0, err
		}

		return validators.Validators, totalValidatorsCount, nil
	} else {
		var count string
		if validators.Result.Total != "" {
			count = validators.Result.Total
		} else {
			count = validators.Result.Pagination.Total
		}
		totalValidatorsCount, err := strconv.ParseInt(count, 10, 64)
		if err != nil {
			return []types.CosmosValidator{}, 0, err
		}

		return validators.Result.Validators, totalValidatorsCount, nil
	}
}

func CosmosValidatorParserGCP(resp []byte) ([]types.CosmosValidator, int64, error) {
	validatorsGCP := &types.CosmosValidatorsGCP{}
	err := json.Unmarshal(resp, validatorsGCP)
	if err != nil {
		return []types.CosmosValidator{}, 0, err
	}
	var count string
	if validatorsGCP.Total != "" {
		count = validatorsGCP.Total
	} else {
		count = validatorsGCP.Pagination.Total
	}
	totalValidatorsCount, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		return []types.CosmosValidator{}, 0, err
	}

	validators := make([]types.CosmosValidator, len(validatorsGCP.Validators))
	for i, v := range validatorsGCP.Validators {

		address, _ := sdkhelper.ProposerAddressFromPublicKey(v.Pubkey.Value)

		validators[i] = types.CosmosValidator{
			Address: address,
			Pubkey: struct {
				Type  string "json:\"type\""
				Value string "json:\"value\""
			}(v.Pubkey),
			VotingPower:      v.VotingPower,
			ProposerPriority: v.ProposerPriority,
		}
	}

	return validators, totalValidatorsCount, nil
}

func CosmosStakingValidatorParser(resp []byte) ([]types.CosmosStakingValidator, error) {
	var result types.CosmosStakingValidatorsQueryResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, common.ErrFailedJsonUnmarshal
	}
	return result.Validators, nil
}

// cosmos upgrade parser
func CosmosUpgradeParser(resp []byte) (
	/* upgrade height */ int64,
	/* upgrade plan name  */ string,
	error) {
	var result types.CosmosUpgradeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, "", fmt.Errorf("parsing error: %s", err.Error())
	}

	if result.Plan.Height == "" {
		return 0, "", nil
	}

	upgradeHeight, err := strconv.ParseInt(result.Plan.Height, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("converting error: %s", err.Error())
	}
	return upgradeHeight, result.Plan.Name, nil
}

func CosmosSlashingParser(resp []byte) (consensusAddress string, indexOffset float64, isTomstoned float64, missedBlocksCounter float64, err error) {
	var result types.CosmosSlashingResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", 0, 0, 0, err
	}
	indexOffset, err = strconv.ParseFloat(result.ValidatorSigningInfo.IndexOffset, 64)
	if err != nil {
		return "", 0, 0, 0, err
	}
	missedBlocksCounter, err = strconv.ParseFloat(result.ValidatorSigningInfo.MissedBlocksCounter, 64)
	if err != nil {
		return "", 0, 0, 0, err
	}

	isTomstoned = float64(0)
	if result.ValidatorSigningInfo.Tombstoned {
		isTomstoned = 1
	}

	return result.ValidatorSigningInfo.ConsensusAddress, indexOffset, isTomstoned, missedBlocksCounter, nil
}

func CosmosSlashingParamsParser(resp []byte) (signedBlocksWindow float64, minSignedPerWindow float64, err error) {
	var result types.CosmosSlashingParamsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, 0, err
	}
	signedBlocksWindow, err = strconv.ParseFloat(result.Params.SignedBlocksWindow, 64)
	if err != nil {
		return 0, 0, err
	}
	minSignedPerWindow, err = strconv.ParseFloat(result.Params.MinSignedPerWindow, 64)
	if err != nil {
		return 0, 0, err
	}
	return signedBlocksWindow, minSignedPerWindow, nil
}
