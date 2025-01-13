package indexer

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pkg/errors"
)

func ParseCurrentEpoch(resp []byte) (int64, int64, error) {
	var currentEpochResponse CurrentEpochResponse
	err := json.Unmarshal(resp, &currentEpochResponse)
	if err != nil {
		return 0, 0, err
	}

	currentEpoch, err := strconv.ParseInt(currentEpochResponse.CurrentEpoch, 10, 64)
	if err != nil {
		return 0, 0, err

	}
	epochBoundaryHeight, err := strconv.ParseInt(currentEpochResponse.EpochBoundary, 10, 64)
	if err != nil {
		return 0, 0, err

	}

	return currentEpoch, epochBoundaryHeight, nil
}

// /* first block height by epoch */ int64,
// /* current epoch interval */ int64,
// /* unexpected error */ error,
func ParseEpoch(resp []byte) (
	/* first block height by epoch */ int64,
	/* current epoch interval */ int64,
	/* unexpected error */ error,
) {
	var result EpochResponse
	err := json.Unmarshal(resp, &result)
	if err != nil {
		return 0, 0, err
	}

	firstBlockHeight, err := strconv.ParseInt(result.Epoch.FirstBlockHeight, 10, 64)
	if err != nil {
		return 0, 0, err

	}
	currentEpochInterval, err := strconv.ParseInt(result.Epoch.CurrentEpochInterval, 10, 64)
	if err != nil {
		return 0, 0, err

	}

	return firstBlockHeight, currentEpochInterval, nil
}

func ParseTxAndExtractBabylonExtendVote(resp []byte) ([]BabylonExtendVote, error) {
	txsReponse := TxsResponse{}
	err := json.Unmarshal(resp, &txsReponse)
	if err != nil {
		return nil, err
	}

	for _, tx := range txsReponse.Txs {
		for _, message := range tx.Body.Messages {
			var preResult map[string]json.RawMessage
			if err := json.Unmarshal(message, &preResult); err != nil {
				return nil, err
			}

			if rawType, ok := preResult["@type"]; ok {
				var typeValue string
				if err := json.Unmarshal(rawType, &typeValue); err != nil {
					return nil, err
				}

				votes, err := ParseDynamicMessage(message, typeValue)
				if err != nil {
					return nil, err
				}

				return votes, nil
			}
		}
	}

	return nil, errors.New("unexpected errors")
}

// parseDynamicMessage dynamically parses the message based on its type.
func ParseDynamicMessage(message json.RawMessage, typeURL string) ([]BabylonExtendVote, error) {
	switch typeURL {
	case BabylonInjectedCheckpointMessageType:
		var msg MsgInjectedCheckpoint
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse MsgInjectedCheckpoint: %v", err)
			return nil, err
		}
		votes := append([]BabylonExtendVote{}, msg.ExtendedCommitInfo.Votes...)
		return votes, nil
	default:
		return nil, fmt.Errorf("unknown message type: %s", typeURL)
	}
}
