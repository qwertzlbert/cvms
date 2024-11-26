package parser

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
)

// band
func BandYodaParser(resp []byte) (float64, error) {
	var result types.BandYodaResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("parsing error: %s", err.Error())
	}
	if !result.Status.IsActive {
		return 0, nil
	}
	return 1, nil
}

func BandYodaParamsParser(resp []byte) (float64, error) {
	var result types.BandYodaParamsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("parsing error: %s", err.Error())
	}
	slashWindow, err := strconv.ParseFloat(result.Params.SlashWindow, 64)
	if err != nil {
		return 0, fmt.Errorf("conversion error: %s", err.Error())
	}
	return slashWindow, nil
}

func BandYodaRequestCountParser(resp []byte) (float64, error) {
	var result types.BandYodaRequestCountResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("parsing error: %s", err.Error())
	}
	requestCount, err := strconv.ParseFloat(result.RequestCount, 64)
	if err != nil {
		return 0, fmt.Errorf("conversion error: %s", err.Error())
	}
	return requestCount, nil
}

func BandLatestBlockParser(resp []byte) (int64, error) {
	var result types.BandLatestBlockHeightResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, fmt.Errorf("parsing error: %s", err.Error())
	}
	latestBlockHeight, err := strconv.ParseInt(result.Block.Header.BlockHeight, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("conversion error: %s", err.Error())
	}
	return latestBlockHeight, nil
}

func BandYodaRequestParser(resp []byte) (
	requestBlock int64,
	validatorsFailedToRespond []string,
	status string,
	err error) {
	var result types.BandYodaRequestResponse
	var requestedValidators []string
	var reporters []string
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, []string{}, "", fmt.Errorf("parsing error: %s", err.Error())
	}
	if result.Request != nil {
		requestBlock, err = strconv.ParseInt(result.Request.RequestBlock, 10, 64)
		if err != nil {
			return 0, []string{}, "", fmt.Errorf("conversion error: %s", err.Error())
		}
		requestedValidators = result.Request.RequestedValidators
	} else {
		requestBlock = 0
		requestedValidators = []string{}
	}
	if result.Result != nil {
		switch result.Result.ResolveStatus {
		case "RESOLVE_STATUS_SUCCESS":
			status = "success"
		default:
			status = "failed"
		}
	} else {
		status = "running"
	}

	for _, reporter := range result.Reports {
		reporters = append(reporters, reporter.Validator)
	}

	for _, validator := range requestedValidators {
		if !(slices.Contains(reporters, validator)) {
			validatorsFailedToRespond = append(validatorsFailedToRespond, validator)
		}
	}
	return requestBlock, validatorsFailedToRespond, status, nil
}
