package api

import (
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestCheckGetBlockResultAndExtractFpVoting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	commonApp := common.NewCommonApp(pBabylon)
	commonApp.SetRPCEndPoint("https://rpc-office.cosmostation.io/babylon-testnet")

	txsEvents, _, _, err := GetBlockResults(commonApp.CommonClient, 92664)
	assert.NoError(t, err)

	const msg = "/babylon.finality.v1.MsgAddFinalitySig"
	for _, e := range txsEvents {
		for _, a := range e.Attributes {
			if a.Value == msg {
				t.Log(a)
				t.Log(e)
			}
		}
	}
}

func Test_Babylon_GetFP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	commonApp := common.NewCommonApp(pBabylon)
	commonApp.SetAPIEndPoint("https://lcd-office.cosmostation.io/babylon-testnet")
	chainID := "bbn-testnet-5"
	fps, err := GetBabylonFinalityProviderInfos(commonApp.CommonClient, chainID)
	assert.NoError(t, err)

	for _, fp := range fps {
		t.Logf("%v", fp)
	}

}
