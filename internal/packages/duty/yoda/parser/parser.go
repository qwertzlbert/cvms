package parser

import (
	"encoding/json"
	"fmt"
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
