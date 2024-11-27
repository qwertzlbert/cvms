package processor_test

import (
	"testing"

	types "github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
	"github.com/stretchr/testify/assert"
)

func TestYodaMissesProcessor(t *testing.T) {

	oldVal1 := types.ValidatorStatus{
		Moniker:                  "testymctestface",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{types.RequestStatus{
			RequestID:                 123,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl", "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9"},
		},
			types.RequestStatus{
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
		Requests: []types.RequestStatus{types.RequestStatus{
			RequestID:                 123,
			Status:                    "running",
			RequestHeight:             1234,
			BlocksPassed:              32,
			ValidatorsFailedToRespond: []string{"bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9", "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl"},
		},
		}}

	newVal1 := types.ValidatorStatus{
		Moniker:                  "testymctestface",
		ValidatorOperatorAddress: "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl",
		IsActive:                 1.0,
		Requests: []types.RequestStatus{types.RequestStatus{
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
		Requests:                 []types.RequestStatus{types.RequestStatus{}},
	}

	validatorResultsOld := []types.ValidatorStatus{oldVal1, oldVal2}
	validatorResultsNew := []types.ValidatorStatus{newVal1, newVal2}

	reqsFinished, error := processor.processYodaMisses(validatorResultsOld, validatorResultsNew)
	assert.Len(t, reqsFinished, 2)

}
