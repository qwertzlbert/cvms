package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cosmostation/cvms/internal/common"
)

// sample rpc event log
// 	{
// 		"key": "covenant_btc_pk_hex",
// 		"value": "\"113c3a32a9d320b72190a04a020a0db3976ef36972673258e9a38a364f3dc3b0\"",
// 		"index": true
// 	}

// value is weird and fixed it
func DecodeEscapedJSONString(value string) (string, error) {
	var decodedString string
	fmt.Println(value)
	err := json.Unmarshal([]byte(value), &decodedString)
	if err != nil {
		return "", err
	}

	return decodedString, nil
}

// {
// 	"@type": "/babylon.btcstaking.v1.MsgAddCovenantSigs",
// 	"signer": "bbn15ewur0qq3fgryz6sappwaqtqk9c45737hznule",
// 	"pk": "79a71ffd71c503ef2e2f91bccfc8fcda7946f4653cef0d9f3dde20795ef3b9f0",
// 	"staking_tx_hash": "3dce373f1142b4b30549a34adec9b9b08c6f7d01ca9fc616b2629bd244c59c43",
// 	"slashing_tx_sigs": [
// 	  "AjI88WdNZPS/eSEOOkBt1FCXKFgAcqQ1RR8bJEHCFTxI6/3vfJfTVZJz5cmOluEIUC2ShcdGL4qUc2ahPbiD/jcA"
// 	],
// 	"unbonding_tx_sig": "o7MLJCZhUtmaJ1vHMVEtzI4TNJ0OpsYgyIFXNA4cka7nnjEsJ62pcRdfzg/XuZJGmFNR+PnA72Mqsa11bnGMqQ==",
// 	"slashing_unbonding_tx_sigs": [
// 	  "AlEOxpq0GMZraZy9tfDp9Tq9KYzgWeqqVN5LTRavU+QHRkKDBCf5GZyKC3GCvnU7azKx9eVtSzz3dSFyV1mIZPIB"
// 	]
//   }

func ExtractBabylonCovenantSignature(resp []byte) (
	/* block height */ int64,
	/* block timestamp */ time.Time,
	[]MsgCovenantSignature,
	error,
) {

	result := BlockTxsResponse{}
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	blockHeight, err := strconv.ParseInt(result.Block.Header.Height, 10, 64)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	newMsgCovenantSigs := make([]MsgCovenantSignature, 0)

	for _, tx := range result.Txs {
		for _, message := range tx.Body.Messages {
			var preResult map[string]json.RawMessage
			if err := json.Unmarshal(message, &preResult); err != nil {
				return 0, time.Time{}, nil, err
			}

			if rawType, ok := preResult["@type"]; ok {
				var typeValue string
				if err := json.Unmarshal(rawType, &typeValue); err != nil {
					return 0, time.Time{}, nil, err
				}

				covenantSigs, err := ParseDynamicMessage(message, typeValue)
				if err != nil {
					if errors.Is(err, common.ErrUnSupportedMessageType) {
						continue
					} else {
						return 0, time.Time{}, nil, err
					}
				}

				newMsgCovenantSigs = append(newMsgCovenantSigs, covenantSigs)
			}
		}
	}

	return blockHeight, result.Block.Header.Time, newMsgCovenantSigs, nil
}

// parseDynamicMessage dynamically parses the message based on its type.
func ParseDynamicMessage(message json.RawMessage, typeURL string) (MsgCovenantSignature, error) {
	switch typeURL {
	case BabylonCovenantSignatureMessageType:
		var msg MsgCovenantSignature
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse MsgCovenantSignature: %v", err)
			return MsgCovenantSignature{}, err
		}
		return msg, nil
	default:
		return MsgCovenantSignature{}, fmt.Errorf("%w", common.ErrUnSupportedMessageType)
	}
}
