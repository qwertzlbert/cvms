package api

import (
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/stretchr/testify/assert"
)

var p = common.Packager{Logger: logger.GetTestLogger()}

func TestCheckGetBlockResultAndExtractFpVoting(t *testing.T) {
	commonApp := common.NewCommonApp(p)
	commonApp.SetRPCEndPoint("https://rpc-office.cosmostation.io/babylon-testnet")

	txsEvents, _, err := GetBlockResults(commonApp.CommonClient, 92664)
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
