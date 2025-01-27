package api

import (
	"context"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/packages/health/block/types"
)

func GetBlockStatus(
	c *common.Exporter,
	CommonBlockCallClient common.ClientType,
	CommonBlockCallMethod common.Method, CommonBlockQueryPath string, CommonBlockPayload string,
	CommonBlockParser func([]byte) (float64, float64, error),
) (types.CommonBlock, error) {
	// init context
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, common.Timeout)
	defer cancel()

	// create requester
	// var requester *resty.Request
	if CommonBlockCallClient == common.RPC {
		// requester = c.RPCClient.R().SetContext(ctx)
	} else {
		// requester = c.APIClient.R().SetContext(ctx)
	}

	// var resp = &resty.Response{}
	var resp []byte
	var err error

	if CommonBlockCallMethod == common.GET {
		resp, err = c.RPCClient.Get(ctx, CommonBlockQueryPath)
		// SetHeader("Content-Type", "application/json").
		// SetBody(CommonBlockPayload).
		// Get(CommonBlockQueryPath)
		// } else if CommonBlockCallMethod == common.POST {
		// 	resp, err = requester.
		// 		SetHeader("Content-Type", "application/json").
		// 		SetBody(CommonBlockPayload).
		// 		Post(CommonBlockQueryPath)
	} else {
		return types.CommonBlock{}, common.ErrUnSupportedMethod
	}

	if err != nil {
		c.Errorf("api error: %s", err)
		return types.CommonBlock{}, common.ErrFailedHttpRequest
	}

	// if resp.StatusCode() != http.StatusOK {
	// 	c.Errorf("request error: [%d] %s", resp.StatusCode(), err)
	// 	return types.CommonBlock{}, common.ErrGotStrangeStatusCode
	// }

	blockHeight, blockTimeStamp, err := CommonBlockParser(resp)
	if err != nil {
		c.Errorf("parser error: %s", err)
		return types.CommonBlock{}, common.ErrFailedJsonUnmarshal
	}

	c.Debugf("got block timestamp: %d", int(blockTimeStamp))
	return types.CommonBlock{
		LastBlockHeight:    blockHeight,
		LastBlockTimeStamp: blockTimeStamp,
	}, nil
}
