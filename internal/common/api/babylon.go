package api

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	fputypes "github.com/cosmostation/cvms/internal/packages/duty/finality-provider-uptime/types"
	"github.com/pkg/errors"
)

func GetActiveFinalityProviderByHeight(c common.CommonClient, height int64) ([]types.FinalityProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	resp, err := requester.Get(types.BabylonFinalityProvidersQueryPath(height))
	if err != nil {
		return nil, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	fps, err := parser.ParseFinalityProviders(resp.Body())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fps.FinalityProviders, nil
}

func GetFinalityVotesByHeight(c common.CommonClient, height int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	resp, err := requester.Get(types.BabylonFinalityVotesQueryPath(height))
	if err != nil {
		return nil, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	votes, err := parser.ParseFinalityProviderVotings(resp.Body())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return votes.BTCPKs, nil
}

func GetBabylonFinalityProviderInfos(c common.CommonClient) ([]types.FinalityProviderInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	fpInfoList := make([]types.FinalityProviderInfo, 0)
	maxCnt := 10
	key := ""
	for cnt := 0; cnt <= maxCnt; cnt++ {
		resp, err := requester.Get(types.BabylonFinalityProviderInfosQueryPath(key))
		if err != nil {
			return nil, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
		}

		if resp.StatusCode() != http.StatusOK {
			return nil, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
		}

		fpInfos, err := parser.ParseFinalityProviderInfos(resp.Body())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		fpInfoList = append(fpInfoList, fpInfos.FinalityProviders...)

		if fpInfos.Pagination.NextKey != "" {
			key = url.QueryEscape(fpInfos.Pagination.NextKey)
			c.Debugf("there is next key, keep collecting finality providers by using this path: %s", types.BabylonFinalityProviderInfosQueryPath(key))
		} else {
			// got all finality provider infos
			c.Debugf("collected all finality providers")
			break
		}
	}

	return fpInfoList, nil
}

func GetFinalityProviderUptime(c common.CommonClient, fpInfoList []types.FinalityProviderInfo) ([]fputypes.FinalityProviderUptimeStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)

	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	validatorResult := make([]fputypes.FinalityProviderUptimeStatus, 0)
	wg.Add(len(fpInfoList))
	for _, item := range fpInfoList {
		moniker := item.Description.Moniker
		Address := item.Address
		BTCPK := item.BTCPK
		queryPath := types.BabylonFinalityProviderSigninInfoQueryPath(item.BTCPK)
		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(c.Entry)
			defer wg.Done()

			resp, err := requester.Get(queryPath)
			if err != nil {
				if resp == nil {
					c.Errorln("unexpected nil response")
					ch <- helper.Result{Item: nil, Success: false}
					return
				}
				c.Errorf("unexpected err: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}
			if resp.StatusCode() != http.StatusOK {
				c.Errorf("unexpected status from node: %d from %s", resp.StatusCode(), resp.Request.URL)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			missedBlockCounter, err := parser.ParserFinalityProviderSigningInfo(resp.Body())
			if err != nil {
				c.Errorf("unexpected err: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			ch <- helper.Result{
				Success: true,
				Item: fputypes.FinalityProviderUptimeStatus{
					Moniker:            moniker,
					Address:            Address,
					BTCPK:              BTCPK,
					MissedBlockCounter: missedBlockCounter,
				}}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	errorCount := 0
	for r := range ch {
		if r.Success {
			validatorResult = append(validatorResult, r.Item.(fputypes.FinalityProviderUptimeStatus))
			continue
		}
		errorCount++
	}

	if errorCount > 0 {
		c.Errorf("current errors count: %d", errorCount)
		return nil, common.ErrFailedHttpRequest
	}

	return validatorResult, nil
}

func GetBabylonFinalityProviderParams(c common.CommonClient) (float64, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	resp, err := requester.Get(types.BabylonFinalityParamsQueryPath)
	if err != nil {
		return 0, 0, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return 0, 0, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	signedBlocksWindow, minSignedPerWindow, err := parser.ParserFinalityParams(resp.Body())
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return signedBlocksWindow, minSignedPerWindow, nil
}
