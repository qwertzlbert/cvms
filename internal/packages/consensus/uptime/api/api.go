package api

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	commonparser "github.com/cosmostation/cvms/internal/common/parser"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	sdkhelper "github.com/cosmostation/cvms/internal/helper/sdk"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/types"
	"github.com/pkg/errors"
)

func GetValidatorUptimeStatus(
	c common.CommonApp,
	CommonUptimeSlashingQueryPath string,
	CommonUptimeSlashingQueryParser func(resp []byte) (
		consensusAddress string,
		indexOffset float64,
		isTomstoned float64,
		missedBlocksCounter float64,
		err error),
	consensusValidators []commontypes.CosmosValidator,
	stakingValidators []commontypes.CosmosStakingValidator) (
	[]types.ValidatorUptimeStatus,
	error) {

	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := c.APIClient.R().SetContext(ctx)

	// extract bech32 valcons prefix using staking validator address
	var bech32ValconsPrefix string
	for idx, validator := range stakingValidators {
		exportedPrefix, err := sdkhelper.ExportBech32ValconsPrefix(validator.OperatorAddress)
		if err != nil {
			return nil, errors.Cause(err)
		}
		if idx == 0 {
			bech32ValconsPrefix = exportedPrefix
			break
		}
	}
	c.Debugf("bech32 valcons prefix: %s", bech32ValconsPrefix)

	// 3. make pubkey map by using consensus hex address with extracted valcons prefix
	pubkeysMap := make(map[string]string)
	for _, validator := range consensusValidators {
		bz, _ := hex.DecodeString(validator.Address)
		consensusAddress, err := sdkhelper.ConvertAndEncode(bech32ValconsPrefix, bz)
		if err != nil {
			return nil, common.ErrFailedConvertTypes
		}
		pubkeysMap[validator.Pubkey.Value] = consensusAddress
	}

	// 4. Sort staking validators by vp
	orderedStakingValidators := sliceStakingValidatorByVP(stakingValidators, len(consensusValidators))

	// 5. init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	validatorResult := make([]types.ValidatorUptimeStatus, 0)
	wg.Add(len(orderedStakingValidators))

	for _, item := range orderedStakingValidators {
		// set query path
		moniker := item.Description.Moniker
		proposerAddress, _ := sdkhelper.ProposerAddressFromPublicKey(item.ConsensusPubkey.Key)
		validatorOperatorAddress := item.OperatorAddress
		consensusAddress := pubkeysMap[item.ConsensusPubkey.Key]
		queryPath := strings.Replace(CommonUptimeSlashingQueryPath, "{consensus_address}", consensusAddress, -1)

		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(c.Entry)
			defer wg.Done()

			resp, err := requester.Get(queryPath)
			if err != nil {
				if resp == nil {
					ch <- helper.Result{Item: nil, Success: false}
					return
				}
				// c.Errorf("errors: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}
			if resp.StatusCode() != http.StatusOK {
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			_, _, isTomstoned, missedBlocksCounter, err := CommonUptimeSlashingQueryParser(resp.Body())
			if err != nil {
				c.Errorf("errors: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			ch <- helper.Result{
				Success: true,
				Item: types.ValidatorUptimeStatus{
					Moniker:                   moniker,
					ProposerAddress:           proposerAddress,
					ValidatorConsensusAddress: consensusAddress,
					MissedBlockCounter:        missedBlocksCounter,
					IsTomstoned:               isTomstoned,
					ValidatorOperatorAddress:  validatorOperatorAddress,
				}}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	// close channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect validator's orch
	errorCount := 0
	for r := range ch {
		if r.Success {
			validatorResult = append(validatorResult, r.Item.(types.ValidatorUptimeStatus))
			continue
		}
		errorCount++
	}

	if errorCount > 0 {
		c.Errorf("current errors count: %d", errorCount)
		return nil, common.ErrFailedHttpRequest
	}
	c.Debugf("got total validator uptime: %d", len(validatorResult))
	return validatorResult, nil
}

func GetUptimeParams(c common.CommonClient, chainName string) (
	/* signed blocks window */ float64,
	/* minimum signed per window */ float64,
	/* unexpected error */ error,
) {
	var (
		queryPath string
		parser    func(resp []byte) (signedBlocksWindow float64, minSignedPerWindow float64, err error)
	)

	switch chainName {
	case "story":
		queryPath = commontypes.StorySlashingParamsQueryPath
		parser = commonparser.StorySlashingParamsParser
	default:
		queryPath = commontypes.CosmosSlashingParamsQueryPath
		parser = commonparser.CosmosSlashingParamsParser
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	resp, err := requester.Get(queryPath)
	if err != nil {
		return 0, 0, errors.Cause(err)
	}
	if resp.StatusCode() != http.StatusOK {
		return 0, 0, errors.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}

	signedBlocksWindow, minSignedPerWindow, err := parser(resp.Body())
	if err != nil {
		return 0, 0, errors.Cause(err)
	}

	c.Debugf("signed block window: %.f", signedBlocksWindow)
	c.Debugf("min signed per window: %.2f", minSignedPerWindow)
	return signedBlocksWindow, minSignedPerWindow, nil
}

// func GetConsumserUptimeStatus(exporter *common.Exporter, chainID string) (types.CommonUptimeStatus, error) {
// 	// set provider client
// 	providerClient := exporter.OptionalClient
// 	consumerClient := exporter.CommonClient

// 	// 1. get consumer id by using chain-id
// 	var consumerID string
// 	consumerChains, err := commonapi.GetConsumerChainID(providerClient)
// 	if err != nil {
// 		return types.CommonUptimeStatus{}, errors.Wrap(err, "failed to get consumer chain id")
// 	}
// 	for _, consumerChain := range consumerChains {
// 		if consumerChain.ChainID == chainID {
// 			consumerID = consumerChain.ConsumerID
// 			break
// 		}
// 	}
// 	// validation check
// 	if consumerID == "" {
// 		return types.CommonUptimeStatus{}, errors.Errorf("failed to find consumer id, check again your chain-id: %s", chainID)
// 	}
// 	exporter.Debugf("got consumer id: %s", consumerID)

// 	// 2. get provider validators
// 	providerStakingValidators, err := commonapi.GetProviderValidators(providerClient, consumerID)
// 	if err != nil {
// 		return types.CommonUptimeStatus{}, errors.Cause(err)
// 	}
// 	exporter.Debugf("got total provider staking validators: %d", len(providerStakingValidators))

// 	// 3. get hrp via slashing info
// 	hrp, err := commonapi.GetConsumerChainHRP(consumerClient)
// 	if err != nil {
// 		return types.CommonUptimeStatus{}, errors.Cause(err)
// 	}
// 	exporter.Debugf("got hrp for making valcons address: %s", hrp)

// 	// 4. get consumer validators uptime status
// 	validatorUptimeStatus, err := getConsumerValidatorUptimeStatus(consumerClient, providerStakingValidators, hrp)
// 	if err != nil {
// 		return types.CommonUptimeStatus{}, errors.Cause(err)
// 	}
// 	exporter.Debugf("got total consumer validator uptime: %d", len(validatorUptimeStatus))

// 	// 5. get on-chain slashing parameter
// 	signedBlocksWindow, minSignedPerWindow, err := getUptimeParams(consumerClient, exporter.ChainName)
// 	if err != nil {
// 		return types.CommonUptimeStatus{}, errors.Cause(err)
// 	}

// 	return types.CommonUptimeStatus{
// 		SignedBlocksWindow: signedBlocksWindow,
// 		MinSignedPerWindow: minSignedPerWindow,
// 		Validators:         validatorUptimeStatus,
// 	}, nil
// }
