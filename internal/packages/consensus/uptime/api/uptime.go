package api

import (
	"sort"
	"strconv"

	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/types"
)

func GetUptimeStatus(signedBlocksWindow float64,
	minSignedPerWindow float64,
	validators []types.ValidatorUptimeStatus) (
	types.CommonUptimeStatus,
	error) {
	return types.CommonUptimeStatus{
		SignedBlocksWindow: signedBlocksWindow,
		MinSignedPerWindow: minSignedPerWindow,
		Validators:         validators,
	}, nil
}

// func getConsumerValidatorUptimeStatus(
// 	app common.CommonClient,
// 	providerValidators []commontypes.ProviderValidator,
// 	consumerValconsPrefix string,
// ) (
// 	[]types.ValidatorUptimeStatus,
// 	error,
// ) {
// 	// init context
// 	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
// 	defer cancel()

// 	// create requester
// 	requester := app.APIClient.R().SetContext(ctx)

// 	// 5. init channel and waitgroup for go-routine
// 	ch := make(chan helper.Result)
// 	var wg sync.WaitGroup
// 	validatorResult := make([]types.ValidatorUptimeStatus, 0)
// 	for _, pv := range providerValidators {
// 		wg.Add(1)
// 		// provider info
// 		moniker := pv.Description.Moniker
// 		providerValoperAddress := pv.ProviderValoperAddress
// 		providerValconsAddress := pv.PrvodierValconsAddress
// 		// consumer info
// 		proposerAddress, _ := sdkhelper.ProposerAddressFromPublicKey(pv.ConsumerKey.Pubkey)
// 		consumerValconsAddress, _ := sdkhelper.MakeValconsAddressFromPubeky(pv.ConsumerKey.Pubkey, consumerValconsPrefix)
// 		uptimeQueryPath := commontypes.CosmosSlashingQueryPath(consumerValconsAddress)

// 		go func(ch chan helper.Result) {
// 			defer helper.HandleOutOfNilResponse(app.Entry)
// 			defer wg.Done()

// 			resp, err := requester.Get(uptimeQueryPath)
// 			if err != nil {
// 				if resp == nil {
// 					ch <- helper.Result{Item: nil, Success: false}
// 					return
// 				}
// 				ch <- helper.Result{Item: nil, Success: false}
// 				return
// 			}
// 			if resp.StatusCode() != http.StatusOK {
// 				ch <- helper.Result{Item: nil, Success: false}
// 				return
// 			}

// 			_, _, isTomstoned, missedBlocksCounter, err := commonparser.CosmosSlashingParser(resp.Body())
// 			if err != nil {
// 				ch <- helper.Result{Item: nil, Success: false}
// 				return
// 			}

// 			ch <- helper.Result{
// 				Success: true,
// 				Item: types.ValidatorUptimeStatus{
// 					// provider
// 					Moniker:                   moniker,
// 					ValidatorConsensusAddress: providerValconsAddress,
// 					ValidatorOperatorAddress:  providerValoperAddress,
// 					IsTomstoned:               isTomstoned,
// 					// consumer
// 					ProposerAddress:          proposerAddress,
// 					ConsumerConsensusAddress: consumerValconsAddress,
// 					MissedBlockCounter:       missedBlocksCounter,
// 				}}
// 		}(ch)
// 		time.Sleep(10 * time.Millisecond)
// 	}

// 	// close channel
// 	go func() {
// 		wg.Wait()
// 		close(ch)
// 	}()

// 	// collect validator's orch
// 	errorCount := 0
// 	for r := range ch {
// 		if r.Success {
// 			validatorResult = append(validatorResult, r.Item.(types.ValidatorUptimeStatus))
// 			continue
// 		}
// 		errorCount++
// 	}

// 	if errorCount > 0 {
// 		app.Errorf("current errors count: %d", errorCount)
// 		return nil, common.ErrFailedHttpRequest
// 	}

// 	return validatorResult, nil
// }

func sliceStakingValidatorByVP(stakingValidators []commontypes.CosmosStakingValidator, totalConsensusValidators int) []commontypes.CosmosStakingValidator {
	sort.Slice(stakingValidators, func(i, j int) bool {
		tokensI, _ := strconv.ParseInt(stakingValidators[i].Tokens, 10, 64)
		tokensJ, _ := strconv.ParseInt(stakingValidators[j].Tokens, 10, 64)
		return tokensI > tokensJ // Sort in descending order
	})
	if len(stakingValidators) < totalConsensusValidators {
		return stakingValidators // Not enough validators to slice
	}
	return stakingValidators[:totalConsensusValidators]
}
