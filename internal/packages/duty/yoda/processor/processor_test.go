package processor_test

import (
	"testing"

	processor "github.com/cosmostation/cvms/internal/packages/duty/yoda/processor"
	types "github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
	"github.com/stretchr/testify/assert"
)

func TestYodaMissesProcessor(t *testing.T) {

	oldVal1 := types.ValidatorStatus{
		Moniker:                  "testymctestface",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{{
			RequestID:                 123,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl", "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9"},
		},
			{
				RequestID:                 234,
				Status:                    "running",
				RequestHeight:             4567,
				BlocksPassed:              64,
				ValidatorsFailedToRespond: []string{"bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl"},
			}},
	}

	oldVal2 := types.ValidatorStatus{
		Moniker:                  "testymctestface2",
		ValidatorOperatorAddress: "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{{
			RequestID:                 123,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9", "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl"},
		},
		}}

	oldVal3 := types.ValidatorStatus{
		Moniker:                  "testymctestface3",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev20",
		IsActive:                 1.0,
		Requests:                 []types.RequestStatus{{}},
	}

	newVal1 := types.ValidatorStatus{
		Moniker:                  "testymctestface",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{{
			RequestID:                 123,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl"},
		},
		}}

	newVal2 := types.ValidatorStatus{
		Moniker:                  "testymctestface2",
		ValidatorOperatorAddress: "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9",
		IsActive:                 1.0,
		Requests:                 []types.RequestStatus{{}},
	}

	newVal3 := types.ValidatorStatus{
		Moniker:                  "testymctestface3",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev20",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{{
			RequestID:                 543,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1xs2penspev20", "xyc"},
		}},
	}

	validatorResultsOld := []types.ValidatorStatus{oldVal1, oldVal2, oldVal3}
	validatorResultsNew := []types.ValidatorStatus{newVal1, newVal2, newVal3}

	reqsFinished := processor.ProcessYodaMisses(validatorResultsOld, validatorResultsNew)
	assert.Len(t, reqsFinished, 2)

	for _, item := range reqsFinished {
		if item.Validator.Moniker == "testymctestface" {
			assert.Equal(t, int64(234), item.Request.RequestID)
		} else if item.Validator.Moniker == "testymctestface2" {
			assert.Equal(t, int64(123), item.Request.RequestID)
		} else if item.Validator.Moniker == "testymctestface3" {
			assert.NotNilf(t, item, "Expected no missed request for validator %s", item.Validator.Moniker)
		}
	}

}
