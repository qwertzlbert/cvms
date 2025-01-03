package api_test

// Import the necessary packages
import (
	"encoding/json"
	"io"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/api"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/jarcoal/httpmock"
)

func setupCommonAppClient() common.CommonApp {

	client := resty.New()
	client.SetBaseURL("https://127.0.0.1")
	l := logger.GetTestLogger()
	restyLogger := logrus.New()
	restyLogger.Out = io.Discard
	entry := l.WithField("mode", "test")
	commonClient := common.CommonClient{
		RPCClient:  client,
		APIClient:  client,
		GRPCClient: client,
		Entry:      entry,
	}
	commonApp := common.CommonApp{
		CommonClient:   commonClient,
		EndPoint:       "https://127.0.0.1",
		OptionalClient: common.CommonClient{},
	}

	return commonApp
}

func TestGetUptimeParams(t *testing.T) {

	commonApp := setupCommonAppClient()
	httpmock.ActivateNonDefault(commonApp.APIClient.GetClient())
	t.Cleanup(httpmock.DeactivateAndReset)

	// mock the response of the uptime params endpoint; Value does not matter for this test
	// as result will be set by the mock query parser
	responder, _ := httpmock.NewJsonResponder(200, json.RawMessage(`{
    	"params": {
    	    "signed_blocks_window": "123",
        	"min_signed_per_window": "0.010000000000000000",
        	"downtime_jail_duration": "60s",
        	"slash_fraction_double_sign": "0.060000000000000000",
        	"slash_fraction_downtime": "0.000100000000000000"
    	}
	}`))

	mockParamsParser := func(resp []byte) (float64, float64, error) {
		return 30000, 0.05, nil
	}

	fakeUrl := "https://127.0.0.1/cosmos/slashing/v1beta1/params"
	httpmock.RegisterResponder("GET", fakeUrl, responder)

	blocksWindow, minSignedBlocks, err := api.GetUptimeParams(
		commonApp,
		"/cosmos/slashing/v1beta1/params",
		mockParamsParser)

	callcount := httpmock.GetTotalCallCount()

	assert.Equal(t, 1, callcount)
	assert.NoError(t, err)
	assert.Equal(t, float64(30000), blocksWindow)
	assert.Equal(t, float64(0.05), minSignedBlocks)

}

// Test the GetValidatorUptimeStatus function
func TestGetValidatorUptimeStatus(t *testing.T) {

	commonApp := setupCommonAppClient()
	httpmock.ActivateNonDefault(commonApp.APIClient.GetClient())
	t.Cleanup(httpmock.DeactivateAndReset)

	mockQueryParser := func(resp []byte) (string, float64, float64, float64, error) {
		return "mockaddress1", 0, 0, 10, nil
	}

	stakingValidators := []commontypes.CosmosStakingValidator{
		{
			OperatorAddress: "cosmosvaloper1exampleaddress",
			ConsensusPubkey: commontypes.ConsensusPubkey{Key: "mockkey1"},
			Description: struct {
				Moniker string "json:\"moniker\""
			}{"validator1"},
		},
	}
	consensusValidators := []commontypes.CosmosValidator{
		{
			Address: "mockaddress1",
			Pubkey: struct {
				Type  string "json:\"type\""
				Value string "json:\"value\""
			}{"ed25519", "mockpubkey1"},
			VotingPower:      "1000",
			ProposerPriority: "0",
		},
	}

	// Mock the response for the uptime endpoint; Value does not matter for this test
	// as result will be set by the mock query parser
	responder, _ := httpmock.NewJsonResponder(200, json.RawMessage(`{
		"asdf": {}
	}`))
	fakeUrl := "https://127.0.0.1/uptime"
	httpmock.RegisterResponder("GET", fakeUrl, responder)

	result, err := api.GetValidatorUptimeStatus(commonApp, "/uptime", mockQueryParser, consensusValidators, stakingValidators)

	callcount := httpmock.GetTotalCallCount()

	assert.NoError(t, err)
	assert.Equal(t, callcount, 1)
	assert.Len(t, result, 1)
	// from consensus validators
	assert.Equal(t, result[0].Moniker, "validator1")
	// from mockQueryParser
	assert.Equal(t, result[0].MissedBlockCounter, float64(10))
	// from staking validators
	assert.Equal(t, result[0].ValidatorOperatorAddress, "cosmosvaloper1exampleaddress")

}
