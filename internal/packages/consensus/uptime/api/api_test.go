package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/api"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementation for Resty request
type MockRestyRequest struct {
	mock.Mock
}

func (r *MockRestyRequest) SetContext(ctx context.Context) *resty.Request {
	r.Called(ctx)
	return &resty.Request{}
}

func (r *MockRestyRequest) Get(url string) (*resty.Response, error) {
	args := r.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

// Mock implementation for Resty client
type MockRestyClient struct {
	mock.Mock
}

func (c *MockRestyClient) R() *resty.Client {
	args := c.Called()
	return args.Get(0).(*resty.Client)
}

func TestGetValidatorUptimeStatus(t *testing.T) {

	validResponse := []byte(`{
		"val_signing_info": {
			"address": "asdf",
			"start_height": "4616678",
			"index_offset": "9554563",
			"jailed_until": "1970-01-01T00:00:00Z",
			"tombstoned": false,
			"missed_blocks_counter": "904"
		}
	}`)

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

	mockAPIClient := new(MockRestyClient)
	mockRequest := new(MockRestyRequest)

	mockAPIClient.On("R").Return(mockRequest)
	mockRequest.On("SetContext", mock.Anything).Return(mockRequest)
	mockRequest.On("Get", mock.Anything).Return(&resty.Response{
		RawResponse: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(validResponse))},
	}, nil)

	// Setup CommonClient and CommonApp
	mockLogger := logrus.NewEntry(logrus.New())
	mockCommonClient := common.CommonClient{
		APIClient: mockAPIClient,
		Entry:     mockLogger,
	}
	mockCommonApp := common.CommonApp{
		CommonClient: mockCommonClient,
		EndPoint:     "http://example.com",
	}

	// Mock query parser
	mockQueryParser := func(resp []byte) (string, float64, float64, float64, error) {
		return "mockaddress1", 0, 0, 10, nil
	}

	// Call the function
	result, err := api.GetValidatorUptimeStatus(
		mockCommonApp,
		"/path/to/query/{consensus_address}",
		mockQueryParser,
		consensusValidators,
		stakingValidators,
	)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

}
