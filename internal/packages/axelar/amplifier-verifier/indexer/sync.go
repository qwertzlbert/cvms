package indexer

import (
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common/api"
	indexermodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/packages/axelar/amplifier-verifier/model"
	"github.com/pkg/errors"
)

func (idx *AxelarAmplifierVerifierIndexer) batchSync(lastIndexPoint int64) (
	/* new index pointer */ int64,
	/* error */ error,
) {
	// set starntHeight and endHeight for batch sync
	// NOTE: end height will use latest height - 10 for block expiry height
	startHeight := (lastIndexPoint + 1)
	endHeight := idx.Lh.LatestHeight
	if startHeight > endHeight {
		idx.Debugf("no need to sync from %d height to %d height, so it'll skip the logic", startHeight, endHeight)
		return lastIndexPoint, nil
	}

	// set limit at end-height in this batch sync logic
	if (endHeight - startHeight) > indexertypes.BatchSyncLimit {
		endHeight = startHeight + indexertypes.BatchSyncLimit
		idx.Debugf("by batch sync limit, end height will change to %d", endHeight)
	}

	// get contract info
	chainNameMap, err := GetVerifierContractAddressMap(idx.CommonClient, false)
	if err != nil {
		return lastIndexPoint, errors.Wrap(err, "failed get verifier register contract address")
	}

	// init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	wg := sync.WaitGroup{}
	summary := make(map[int64]PollDataSummary)

	// start to call block results
	for h := startHeight; h <= endHeight; h++ {
		wg.Add(1)
		height := h

		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(idx.Entry)
			defer wg.Done()

			txsEvents, _, err := api.GetBlockResults(idx.CommonClient, height)
			if err != nil {
				idx.Errorf("failed to call at %d height block results, %s", height, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			if len(txsEvents) == 0 {
				polls := AmplifierPollStartFillter(txsEvents)
				ch <- helper.Result{
					Item: PollDataSummary{
						height:    height,
						polls:     polls,
						pollVotes: nil,
					},
					Success: true,
				}
				return
			}

			_, _, txs, err := api.GetBlockAndTxs(idx.CommonClient, height)
			if err != nil {
				idx.Errorf("failed to call at %d height block and txs, %s", height, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			pollVotes, err := ExtractPoll(txs)
			if err != nil {
				idx.Errorln(err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			polls := AmplifierPollStartFillter(txsEvents)
			ch <- helper.Result{
				Item: PollDataSummary{
					height:    height,
					polls:     polls,
					pollVotes: pollVotes,
				},
				Success: true,
			}

			idx.Debugf("got poll: %d and votes: %d in %d", len(polls), len(pollVotes), height)
		}(ch)

		time.Sleep(10 * time.Millisecond)
	}

	// close channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect block summary data into block summary list
	errorCount := 0
	for r := range ch {
		if r.Success {
			item := r.Item.(PollDataSummary)
			summary[item.height] = item
			continue
		}
		errorCount++
	}

	// check error count
	if errorCount > 0 {
		return lastIndexPoint, errors.Errorf("failed to collect batch poll data, total errors: %d", errorCount)
	}

	// first add new verifiers
	isNewVerifier := false
	newVerifierInfo := make([]indexermodel.VerifierInfo, 0)
	for h := startHeight; h <= endHeight; h++ {
		for _, poll := range summary[h].polls {
			for _, verifier := range poll.Participants {
				_, exist := idx.Vim[verifier]
				if !exist {
					idx.Debugf("the %s isn't in current verifier info table, the address will be added into the meta table", verifier)
					isNewVerifier = true
					newVerifierInfo = append(newVerifierInfo, indexermodel.VerifierInfo{
						ChainInfoID:     idx.ChainInfoID,
						VerifierAddress: verifier,
						Moniker:         verifier,
					})
				}
			}
		}
	}

	if isNewVerifier {
		idx.Debugf("insert new amplifier verifiers: %d", len(newVerifierInfo))
		err := idx.InsertVerifierInfoList(newVerifierInfo)
		if err != nil {
			// NOTE: fetch again validator_info list, actually already inserted the list by other indexer service
			idx.FetchValidatorInfoList()
			return lastIndexPoint, errors.Wrap(err, "failed to insert new reporter list")
		}

		verifierInfoList, err := idx.GetVerifierInfoListByChainInfoID(idx.ChainInfoID)
		if err != nil {
			return lastIndexPoint, errors.Wrap(err, "failed to get new reporter info list after inserting new hex address list")

		}

		for _, v := range verifierInfoList {
			idx.Vim[v.VerifierAddress] = int64(v.ID)
			idx.VAM[v.ID] = v.VerifierAddress
		}

		idx.Debugf("changed vim length: %d and VAM: %d", len(idx.Vim), len(idx.VAM))
	}

	initPollVoteList := make([]model.AxelarAmplifierVerifierVote, 0)
	pollVoteList := make([]model.AxelarAmplifierVerifierVote, 0)
	polls := make([]Poll, 0)
	for h := startHeight; h <= endHeight; h++ {
		for _, poll := range summary[h].polls {
			chainAndPollID := ConcatChainAndPollID(poll.SourceChain, poll.PollID)
			for _, verifier := range poll.Participants {
				initPollVoteList = append(initPollVoteList, model.AxelarAmplifierVerifierVote{
					// ID: Autoincrement
					ChainInfoID:     idx.ChainInfoID,
					CreatedAt:       time.Now(),
					ChainAndPollID:  chainAndPollID,
					PollStartHeight: h,
					PollVoteHeight:  0,
					VerifierID:      idx.Vim[verifier],
					Status:          model.PollStart,
				})
			}
			idx.Debugf("%s was inited", chainAndPollID)

			polls = append(polls, poll)
		}

		for _, pv := range summary[h].pollVotes {
			contractInfo, exist := chainNameMap[pv.ContractAddress]
			if !exist {
				return lastIndexPoint, errors.Wrap(err, "unexpected poll voted was occured")
			}

			verifierID, exist := idx.Vim[pv.VerifierAddress]
			if !exist {
				return lastIndexPoint, errors.Wrap(err, "unexpected verifier address existed")
			}

			// NOTE: if height is already over block expiry, we need to change status from success to did not vote..
			pollVoteList = append(pollVoteList, model.AxelarAmplifierVerifierVote{
				// where
				ChainInfoID:    idx.ChainInfoID,
				ChainAndPollID: ConcatChainAndPollID(contractInfo.ChainName, pv.PollID),
				VerifierID:     verifierID,
				// set
				Status:         model.StringToPollVote(pv.StatusStr),
				PollVoteHeight: h,
			})

		}
		idx.Debugf("%d poll votes will be updated in %d height", len(summary[h].pollVotes), h)
	}

	// 1. insert init votes
	if len(initPollVoteList) > 0 {
		err = idx.InsertInitPollVoteList(idx.ChainInfoID, initPollVoteList)
		if err != nil {
			return lastIndexPoint, errors.Wrap(err, "failed to insert init vote list")
		}
	}

	// 2. update poll votes
	err = idx.UpdatePollVoteList(idx.ChainInfoID, endHeight, pollVoteList)
	if err != nil {
		return lastIndexPoint, errors.Wrap(err, "failed to insert update poll vote list")
	}

	// 3. update metrics
	idx.updatePrometheusMetrics(endHeight, polls)
	idx.updatePollVoteStatusMetric()

	return endHeight, nil
}

type PollDataSummary struct {
	height    int64
	polls     []Poll
	pollVotes []PollVote
}

// func (idx *AxelarAmplifierVerifierIndexer) MustInitPoll(chainAndPollID, contractAddress, pollID string, blockExpiry int64) map[string]model.AxelarAmplifierVerifierVote {
// 	expireHeight, participants, err := GetPollState(idx.CommonClient, contractAddress, pollID)
// 	if err != nil {
// 		return nil
// 	}

// 	isNewVerifier := false
// 	newVerifierInfo := make([]indexermodel.VerifierInfo, 0)
// 	for _, verifier := range participants {
// 		_, exist := idx.Vim[verifier]
// 		if !exist {
// 			idx.Debugf("the %s isn't in current verifier info table, the address will be added into the meta table", verifier)
// 			isNewVerifier = true
// 			newVerifierInfo = append(newVerifierInfo, indexermodel.VerifierInfo{
// 				ChainInfoID:     idx.ChainInfoID,
// 				VerifierAddress: verifier,
// 				Moniker:         verifier,
// 			})
// 		}
// 	}

// 	if isNewVerifier {
// 		idx.Debugf("insert new amplifier verifiers: %d", len(newVerifierInfo))
// 		err := idx.InsertVerifierInfoList(newVerifierInfo)
// 		if err != nil {
// 			return nil
// 		}

// 		err = idx.FetchValidatorInfoList()
// 		if err != nil {
// 			return nil
// 		}

// 		verifierInfoList, err := idx.GetVerifierInfoListByChainInfoID(idx.ChainInfoID)
// 		if err != nil {
// 			return nil
// 		}

// 		for _, v := range verifierInfoList {
// 			idx.Vim[v.VerifierAddress] = int64(v.ID)
// 			idx.VAM[v.ID] = v.VerifierAddress
// 		}
// 	}

// 	initVoteMap := make(map[string]model.AxelarAmplifierVerifierVote, len(participants))
// 	pollStartHeight := expireHeight - blockExpiry
// 	for _, verifier := range participants {
// 		initVote := model.AxelarAmplifierVerifierVote{
// 			// ID: Autoincrement
// 			ChainInfoID:     idx.ChainInfoID,
// 			CreatedAt:       time.Now(),
// 			ChainAndPollID:  chainAndPollID,
// 			PollStartHeight: pollStartHeight,
// 			PollVoteHeight:  0,
// 			VerifierID:      idx.Vim[verifier],
// 			Status:          model.PollStart,
// 		}
// 		initVoteMap[verifier] = initVote
// 	}

// 	return initVoteMap
// }
