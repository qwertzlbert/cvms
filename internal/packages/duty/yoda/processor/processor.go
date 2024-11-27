package processor

import (
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
)

func ProcessYodaMisses(oldValidatorStatus []types.ValidatorStatus, newValidatorStatus []types.ValidatorStatus) []types.MissedRequests {
	// This function compares two slices of validator statuses to find the differences between the old and new validator statuses.
	// The difference is a list of finished requests per validator.
	// By only counting request misses which are no longer part of newValidatorStatus
	// current slice double counting is prevented

	var reqsFinished []types.MissedRequests

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
				if _, exists := newReqIDs[oldReq.RequestID]; !exists && oldReq.RequestID > 0 {
					// Record the miss
					reqsFinished = append(reqsFinished, types.MissedRequests{Validator: oldVal, Request: oldReq})
				}
			}
		}
	}

	return reqsFinished // Placeholder return statement
}
