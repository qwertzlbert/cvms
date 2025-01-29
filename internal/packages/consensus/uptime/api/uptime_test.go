package api

import (
	"testing"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/client"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper/logger"

	"github.com/stretchr/testify/assert"
)

func TestGetValidatorUptimeStatus(t *testing.T) {

	goodResults := make(map[string][]byte)
	goodResults["/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons1m8u2gxmc92n2v6kus8u48y3u0h88kcqp69cr3a"] = []byte(`{
    "val_signing_info": {
        "address": "cosmosvalcons1m8u2gxmc92n2v6kus8u48y3u0h88kcqp69cr3a",
        "start_height": "0",
        "index_offset": "27266009",
        "jailed_until": "1970-01-01T00:00:00Z",
        "tombstoned": false,
        "missed_blocks_counter": "39"
    }
}`)

	goodResults["/cosmos/slashing/v1beta1/signing_infos/"] = []byte(`{
    "info": [
        {
            "address": "cosmosvalcons1qqqqrezrl53hujmpdch6d805ac75n220ku09rl",
            "start_height": "0",
            "index_offset": "27266009",
            "jailed_until": "1970-01-01T00:00:00Z",
            "tombstoned": false,
            "missed_blocks_counter": "1"
        },
        {
            "address": "cosmosvalcons1m8u2gxmc92n2v6kus8u48y3u0h88kcqp69cr3a",
            "start_height": "0",
            "index_offset": "27550724",
            "jailed_until": "1970-01-01T00:00:00Z",
            "tombstoned": false,
            "missed_blocks_counter": "39"
        }],
		    "pagination": {
        "total": "2"
    }
}`)

	validators := []commontypes.CosmosValidator{
		{
			Address: "D9F8A41B782AA6A66ADC81F953923C7DCE7B6001",
			Pubkey: struct {
				Type  string "json:\"type\""
				Value string "json:\"value\""
			}{
				Type:  "tendermint/PubKeyEd25519",
				Value: "CiCf1tb2Ga/BLCb8Ccm0PBk0YD7ZFKhCx+gwWxzZGtvS9g==",
			},
			VotingPower:      "1",
			ProposerPriority: "0",
		},
	}

	stakingValidators := []commontypes.CosmosStakingValidator{
		{
			OperatorAddress: "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d",
			ConsensusPubkey: commontypes.ConsensusPubkey{
				Type: "/cosmos.crypto.ed25519.PubKey",
				Key:  "CiCf1tb2Ga/BLCb8Ccm0PBk0YD7ZFKhCx+gwWxzZGtvS9g==",
			},
			Description: struct {
				Moniker string "json:\"moniker\""
			}{
				Moniker: "asdf",
			},
			Tokens: "1000",
		},
	}

	l := logger.GetTestLogger()

	commonApp := common.CommonApp{

		CommonClient: common.CommonClient{
			RPCClient:  nil,
			GRPCClient: nil,
			APIClient:  client.NewMockClient("127.0.0.1:9090", goodResults).SetLogger(l),
			Entry:      l.WithField("mode", "test"),
		},
		EndPoint:       "127.0.0.1:9090",
		OptionalClient: common.CommonClient{},
	}

	status, _ := getValidatorUptimeStatus(commonApp, "cosmoshub", validators, stakingValidators)

	assert.Equal(t, float64(39), status[0].MissedBlockCounter)
	assert.Equal(t, float64(0), status[0].IsTomstoned)

}

func TestGetUptimeParamsHappy(t *testing.T) {

	goodResults := make(map[string][]byte)
	goodResults["/cosmos/slashing/v1beta1/params"] = []byte(`{
		"params": {
			"signed_blocks_window": "10000",
			"min_signed_per_window": "0.050000000000000000",
			"downtime_jail_duration": "600s",
			"slash_fraction_double_sign": "0.050000000000000000",
			"slash_fraction_downtime": "0.000100000000000000"
		}
	}`)

	l := logger.GetTestLogger()

	commonClient := common.CommonClient{
		RPCClient:  nil,
		GRPCClient: nil,
		APIClient:  client.NewMockClient("127.0.0.1:9090", goodResults).SetLogger(l),
		Entry:      l.WithField("mode", "test"),
	}

	signedBlocksWindow, minSignedPerWindow, downtimeJailDuration, slashFractionDowntime, slashFractionDoubleSign, err := getUptimeParams(commonClient, "chainyMCChainface")
	assert.Nil(t, err)
	assert.Equal(t, float64(10000), signedBlocksWindow)
	assert.Equal(t, float64(0.05), minSignedPerWindow)
	assert.Equal(t, time.Minute*10, downtimeJailDuration)
	assert.Equal(t, float64(0.0001), slashFractionDowntime)
	assert.Equal(t, float64(0.05), slashFractionDoubleSign)
}

func TestGetUptimeParamsBadResponse(t *testing.T) {
	badResult := make(map[string][]byte)
	badResult["/cosmos/slashing/v1beta1/params"] = []byte(`ERROR: UNEXPECTED RESPONSE`)

	l := logger.GetTestLogger()

	commonClient := common.CommonClient{
		RPCClient:  nil,
		GRPCClient: nil,
		APIClient:  client.NewMockClient("127.0.0.1:9090", badResult).SetLogger(l),
		Entry:      l.WithField("mode", "test"),
	}

	_, _, _, _, _, err := getUptimeParams(commonClient, "chainyMCChainface")
	assert.Error(t, err)
}
