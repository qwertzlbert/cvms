package api

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper/logger"
)

var (
	pBabylon = common.Packager{
		ChainName:    "babylon",
		ChainID:      "bbn-test-5",
		ProtocolType: "cosmos",
		Endpoints: common.Endpoints{
			RPCs: []string{"https://babylon-testnet-rpc.polkachu.com"},
			APIs: []string{"https://lcd-office.cosmostation.io/babylon-testnet"},
		},
		Logger: logger.GetTestLogger(),
	}
	pGnoland = common.Packager{
		ChainName:    "gnoland",
		ChainID:      "test6",
		ProtocolType: "gnoland",
		Endpoints: common.Endpoints{
			RPCs: []string{"https://rpc.test6.testnets.gno.land"},
			APIs: []string{},
		},
		Logger: logger.GetTestLogger(),
	}
)
