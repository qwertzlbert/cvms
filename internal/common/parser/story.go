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

func StoryStakingValidatorParser(resp []byte) ([]types.CosmosStakingValidator, error) {
	var result types.StoryStakingValidatorsQueryResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, common.ErrFailedJsonUnmarshal
	}

	stakingValidatorList := make([]types.CosmosStakingValidator, 0)
	for _, validator := range result.Msg.Validators {
		// const Secp256k1 = "/cosmos.crypto.secp256k1.PubKey"
		// const TendermintSecp256k1 = "tendermint/PubKeySecp256k1"
		if validator.ConsensusPubkey.Type != sdkhelper.TendermintSecp256k1 {
			return nil, errors.New("unexpected key type for story validators")
		}

		stakingValidatorList = append(stakingValidatorList, types.CosmosStakingValidator{
			OperatorAddress: validator.OperatorAddress,
			Description:     validator.Description,
			// story not same consensus pubkey result.
			ConsensusPubkey: types.ConsensusPubkey{
				Type: sdkhelper.Secp256k1,
				Key:  validator.ConsensusPubkey.Value,
			},
		})
	}

	return stakingValidatorList, nil
}

// story upgrade parser
func StoryUpgradeParser(resp []byte) (
	/* upgrade height */ int64,
	/* upgrade plan name  */ string,
	error) {
	var result types.StoryUpgradeResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, "", fmt.Errorf("parsing error: %s", err.Error())
	}

	if result.Msg.Plan.Height == "" {
		return 0, "", nil
	}

	upgradeHeight, err := strconv.ParseInt(result.Msg.Plan.Height, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("converting error: %s", err.Error())
	}
	return upgradeHeight, result.Msg.Plan.Name, nil
}

// story slashing parser
func StorySlashingParser(resp []byte) (consensusAddress string, indexOffset float64, isTomstoned float64, missedBlocksCounter float64, err error) {
	var result types.StorySlashingResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", 0, 0, 0, err
	}
	indexOffset, err = strconv.ParseFloat(result.Msg.ValidatorSigningInfo.IndexOffset, 64)
	if err != nil {
		return "", 0, 0, 0, errors.Wrap(err, "no index_offset key in the response")
	}

	if result.Msg.ValidatorSigningInfo.MissedBlocksCounter != "" {
		missedBlocksCounter, err = strconv.ParseFloat(result.Msg.ValidatorSigningInfo.MissedBlocksCounter, 64)
		if err != nil {
			return "", 0, 0, 0, err
		}
	}

	isTomstoned = float64(0)
	if result.Msg.ValidatorSigningInfo.Tombstoned {
		isTomstoned = 1
	}

	return result.Msg.ValidatorSigningInfo.ConsensusAddress, indexOffset, isTomstoned, missedBlocksCounter, nil
}

func StorySlashingParamsParser(resp []byte) (signedBlocksWindow float64,
	minSignedPerWindow float64,
	downtimeJailDuration time.Duration,
	slashFractionDowntime float64,
	slashFractionDoubleSign float64,
	err error) {
	var result types.StorySlashingParamsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, 0, 0, 0, 0, err
	}
	signedBlocksWindow, err = strconv.ParseFloat(result.Msg.Params.SignedBlocksWindow, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	minSignedPerWindow, err = strconv.ParseFloat(result.Msg.Params.MinSignedPerWindow, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	downtimeJailDuration, err = time.ParseDuration(result.Msg.Params.DowntimeJailDuration)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	slashFractionDowntime, err = strconv.ParseFloat(result.Msg.Params.SlashFractionDowntime, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	slashFractionDoubleSign, err = strconv.ParseFloat(result.Msg.Params.SlashFractionDoubleSign, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return signedBlocksWindow, minSignedPerWindow, downtimeJailDuration, slashFractionDowntime, slashFractionDoubleSign, nil
}
