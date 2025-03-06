package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/packages/duty/axelar-evm/parser"
	"github.com/cosmostation/cvms/internal/packages/duty/axelar-evm/types"
)

var latestHeartbeatsHeight int64 = 0 // global state for heartbeats epoch

func GetAxelarNexusStatus(
	exporter *common.Exporter,
	CommonEvmChainsQueryPath string,
	CommonEvmChainsParser func([]byte) ([]string, error),
	CommonEvmChainMaintainerQueryPath string,
	CommonChainMaintainersParser func([]byte) ([]string, error),
) (types.CommonAxelarNexus, error) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := exporter.APIClient.R().SetContext(ctx)

	// get on-chain validators
	resp, err := requester.Get(types.CommonValidatorQueryPath)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarNexus{}, common.ErrFailedHttpRequest
	}
	if resp.StatusCode() != http.StatusOK {
		exporter.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
		return types.CommonAxelarNexus{}, common.ErrGotStrangeStatusCode
	}

	// json unmarsharling received validators data
	var validators types.CommonValidatorsQueryResponse
	if err := json.Unmarshal(resp.Body(), &validators); err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarNexus{}, common.ErrFailedJsonUnmarshal
	}

	// get on-chain active evm-chains
	resp, err = requester.Get(CommonEvmChainsQueryPath)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarNexus{}, common.ErrFailedHttpRequest
	}
	if resp.StatusCode() != http.StatusOK {
		exporter.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
		return types.CommonAxelarNexus{}, common.ErrGotStrangeStatusCode
	}

	activatedEvmChains, err := CommonEvmChainsParser(resp.Body())
	if err != nil {
		return types.CommonAxelarNexus{}, err
	}
	exporter.Debugln("currently activated evm chains in axelar:", activatedEvmChains)

	// init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	totalStatus := make([]types.ValidatorStatus, 0)

	// add wg by the number of active evm chains
	wg.Add(len(activatedEvmChains))

	// get evm status by each validator
	for _, evmChain := range activatedEvmChains {
		// set query path and variables
		queryPath := strings.Replace(CommonEvmChainMaintainerQueryPath, "{chain}", evmChain, -1)
		maintainerMap := make(map[string]float64)
		chainStatus := make([]types.ValidatorStatus, 0)
		chainName := evmChain

		// start go-routine
		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(exporter.Entry)
			defer wg.Done()

			resp, err = requester.Get(queryPath)
			if err != nil {
				if resp == nil {
					exporter.Errorln("[panic] passed resp.Time() nil point err")
					ch <- helper.Result{Item: nil, Success: false}
					return
				}
				exporter.Errorf("api error: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}
			if resp.StatusCode() != http.StatusOK {
				exporter.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			maintainers, err := CommonChainMaintainersParser(resp.Body())
			if err != nil {
				exporter.Errorf("api error: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			for _, maintainer := range maintainers {
				maintainerMap[maintainer] = 1
			}

			for _, item := range validators.Validators {
				chainStatus = append(chainStatus, types.ValidatorStatus{
					Moniker:                  item.Description.Moniker,
					ValidatorOperatorAddress: item.OperatorAddress,
					Status:                   maintainerMap[item.OperatorAddress],
					EVMChainName:             chainName,
				})
			}

			exporter.Debugf("total validators: %d and %s chain status results: %d", len(validators.Validators), chainName, len(chainStatus))
			ch <- helper.Result{
				Success: true,
				Item:    chainStatus,
			}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	// close channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect validator's orch
	errorCounter := 0
	for r := range ch {
		if r.Success {
			if item, ok := r.Item.([]types.ValidatorStatus); ok {
				totalStatus = append(totalStatus, item...)
				continue
			}
		}
		errorCounter++
	}

	if errorCounter > 0 {
		return types.CommonAxelarNexus{}, errors.New("failed to get all validators status from go-routine")
	}

	return types.CommonAxelarNexus{
		ActiveEVMChains: activatedEvmChains,
		Validators:      totalStatus,
	}, nil
}

func GetAxelarHeartbeatsStatus(
	exporter *common.Exporter,
	CommonProxyResisterQueryPath string,
	CommonProxyResisterParser func([]byte) (types.AxelarProxyResisterStatus, error),
) (types.CommonAxelarHeartbeats, error) {
	currentBlockHeight, _, err := api.GetStatus(exporter.CommonClient)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedHttpRequest
	}

	//Always 51,101,151,201... submit heartbeat tx
	var findHeartbeatsHeight int64
	if currentBlockHeight%50 != 0 {
		findHeartbeatsHeight = currentBlockHeight - (currentBlockHeight % 50) + 1
	} else {
		findHeartbeatsHeight = currentBlockHeight - 50 + 1
	}

	// Skip if heartbeatsHeight is greater than or equal to
	if latestHeartbeatsHeight >= findHeartbeatsHeight {
		return types.CommonAxelarHeartbeats{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := exporter.APIClient.R().SetContext(ctx)

	// get on-chain validators
	resp, err := requester.Get(types.CommonValidatorQueryPath)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedHttpRequest
	}
	if resp.StatusCode() != http.StatusOK {
		exporter.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
		return types.CommonAxelarHeartbeats{}, common.ErrGotStrangeStatusCode
	}

	// json unmarsharling received validators data
	var validators types.CommonValidatorsQueryResponse
	if err := json.Unmarshal(resp.Body(), &validators); err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedJsonUnmarshal
	}

	// // init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	totalBroadcastorStatus := make([]types.BroadcastorStatus, 0)

	updateHeight, err := findHeartbeats(ctx, &ch, &wg, exporter, validators, findHeartbeatsHeight)

	// close channel
	errorCounter := 0
	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		if r.Success {
			if item, ok := r.Item.(types.BroadcastorStatus); ok {
				totalBroadcastorStatus = append(totalBroadcastorStatus, item)
				continue
			}
		}
		errorCounter++
	}

	if errorCounter > 0 {
		return types.CommonAxelarHeartbeats{}, errors.New("failed to get heartbeats from go-routine")
	}

	latestHeartbeatsHeight = updateHeight
	return types.CommonAxelarHeartbeats{Validators: totalBroadcastorStatus}, nil
}

func findHeartbeats(
	ctx context.Context,
	ch *chan helper.Result,
	wg *sync.WaitGroup,
	exporter *common.Exporter,
	validators types.CommonValidatorsQueryResponse,
	heartbeatsHeight int64,
) (int64, error) {

	tryCount := 2
	blockTxsCache := make([]commontypes.CosmosTx, 0)

	// Sometimes, due to block size or timing,
	// the heartbeats tx gets put into the next block,
	// so we check +1 block more to find it.
	for i := 0; i < tryCount; i++ {
		blockHeight := heartbeatsHeight + int64(i)

		_, _, blockTxs, err := api.GetBlockAndTxs(exporter.CommonClient, blockHeight)
		if err != nil {
			exporter.Errorf("API error: %s", err)
			return heartbeatsHeight, common.ErrFailedHttpRequest
		}

		blockTxsCache = append(blockTxsCache, blockTxs...)
	}

	for _, validator := range validators.Validators {
		wg.Add(1)
		operatorAddr := validator.OperatorAddress
		moniker := validator.Description.Moniker

		go func(operatorAddr, moniker string) {
			defer wg.Done()

			// Find Broadcastor address
			rpcRequester := exporter.RPCClient.R().SetContext(ctx)
			abciQueryPath := strings.Replace(types.AxelarProxyResisterQueryPath, "{validator_operator_address}", operatorAddr, -1)
			resp, err := rpcRequester.Get(abciQueryPath)
			if err != nil {
				exporter.Errorf("API error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var AxelarProxyResisterResponse types.AxelarProxyResisterResponse
			if err := json.Unmarshal(resp.Body(), &AxelarProxyResisterResponse); err != nil {
				exporter.Errorf("JSON unmarshal error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			// The ABCI query returns an encoded value, so try to decode it to base64.
			decodedBytes, err := base64.StdEncoding.DecodeString(AxelarProxyResisterResponse.Result.Response.Value)
			if err != nil {
				exporter.Errorf("Failed to decode Base64: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var AxelarProxyResisterStatus types.AxelarProxyResisterStatus
			if err := json.Unmarshal(decodedBytes, &AxelarProxyResisterStatus); err != nil {
				exporter.Errorf("JSON unmarshal error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			newBroadcastorStatus := types.BroadcastorStatus{
				Moniker:                  moniker,
				ValidatorOperatorAddress: operatorAddr,
				BroadcastorAddress:       AxelarProxyResisterStatus.Address,
				Status:                   "",
			}

			foundHealthyTx := false

			for i := 0; i < tryCount; i++ {
				for _, tx := range blockTxsCache {
					isHealthy, err := parser.AxelarHeartbeatsFilterInTx(tx, AxelarProxyResisterStatus.Address)
					if err != nil {
						exporter.Error(err)
						*ch <- helper.Result{Item: nil, Success: false}
						return
					}

					if isHealthy {
						newBroadcastorStatus.Status = "success"
						foundHealthyTx = true
						break
					}
					if foundHealthyTx {
						break
					}
				}
			}

			// If heartbeats tx is not found after scanning all blocks, handle missed
			if !foundHealthyTx {
				newBroadcastorStatus.Status = "missed"
			}

			*ch <- helper.Result{Item: newBroadcastorStatus, Success: true}
		}(operatorAddr, moniker)
	}

	return heartbeatsHeight + int64(tryCount-1), nil
}
