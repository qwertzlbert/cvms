package processor

import (
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
)

func processYodaMisses(oldValidatorStatus []types.ValidatorStatus, newValidatorStatus []types.ValidatorStatus) ([]map[*types.ValidatorStatus]map[*types.RequestStatus]struct{}, error) {

	newRequestsMap := make(map[string]map[int64]struct{})
	for _, newVal := range newValidatorStatus {
		reqIDs := make(map[int64]struct{})
		for _, req := range newVal.Requests {
			reqIDs[req.RequestID] = struct{}{}
		}
		newRequestsMap[newVal.ValidatorOperatorAddress] = reqIDs
	}

	// Iterate through old validators and compare
	for _, oldVal := range oldValidatorStatus {
		if newReqIDs, found := newRequestsMap[oldVal.ValidatorOperatorAddress]; found {
			for _, oldReq := range oldVal.Requests {
				if _, exists := newReqIDs[oldReq.RequestID]; !exists {
					// Record the miss
					return nil, nil
				}
			}
		}
	}

	return nil, nil // Placeholder return statement
}
