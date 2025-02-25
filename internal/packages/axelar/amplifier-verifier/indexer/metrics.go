package indexer

import (
	"strings"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	PollMetricName   = "poll"
	PollLabel        = "poll_id"
	SourceChainLabel = "source_chain"

	PollVoteMetricName = "poll_vote"
	VoteStatusLabel    = "status"
	VerifierLabel      = "verifier"
)

func (idx *AxelarAmplifierVerifierIndexer) initLabelsAndMetrics() {
	idx.MetricsMap[common.IndexPointerBlockHeightMetricName] = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockHeightMetricName,
		ConstLabels: idx.PackageLabels,
	})

	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName] = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockTimestampMetricName,
		ConstLabels: idx.PackageLabels,
	})

	latestBlockHeightMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.LatestBlockHeightMetricName,
		ConstLabels: idx.PackageLabels,
	})
	latestBlockHeightMetric.Set(0)
	idx.MetricsMap[common.LatestBlockHeightMetricName] = latestBlockHeightMetric

	// only axelar amplifier verifier
	idx.MetricsCountVecMap[PollMetricName] = idx.Factory.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   common.Namespace,
			Subsystem:   subsystem,
			Name:        PollMetricName,
			ConstLabels: idx.PackageLabels},
		[]string{
			PollLabel,
			SourceChainLabel,
		})

	idx.MetricsVecMap[PollVoteMetricName] = idx.Factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   common.Namespace,
			Subsystem:   subsystem,
			Name:        PollVoteMetricName,
			ConstLabels: idx.PackageLabels},
		[]string{
			VoteStatusLabel,
			VerifierLabel,
		})
}

func (idx *AxelarAmplifierVerifierIndexer) updatePrometheusMetrics(indexPointer int64, polls []Poll) {
	for _, poll := range polls {
		idx.MetricsCountVecMap[PollMetricName].
			With(prometheus.Labels{
				PollLabel:        strings.ReplaceAll(poll.PollID, `"`, ``),
				SourceChainLabel: poll.SourceChain,
			}).
			Inc()
	}
	idx.MetricsMap[common.IndexPointerBlockHeightMetricName].Set(float64(indexPointer))
	_, timestamp, _, _, _, _, err := api.GetBlock(idx.CommonClient, indexPointer)
	if err != nil {
		idx.Errorf("failed to get block %d: %s", indexPointer, err)
		return
	}
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName].Set((float64(timestamp.Unix())))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}

func (idx *AxelarAmplifierVerifierIndexer) updatePollVoteStatusMetric() {
	pollVoteList, err := idx.SelectPollVoteStatus(idx.ChainID)
	if err != nil {
		idx.Errorf("failed to select poll vote status: %s", err)
	}

	for _, pv := range pollVoteList {
		idx.MetricsVecMap[PollVoteMetricName].
			With(prometheus.Labels{
				VoteStatusLabel: "DidNotVote",
				VerifierLabel:   pv.Moniker,
			}).
			Set(float64(pv.DidNotVote))

		idx.MetricsVecMap[PollVoteMetricName].
			With(prometheus.Labels{
				VoteStatusLabel: "FailedOnChain",
				VerifierLabel:   pv.Moniker,
			}).
			Set(float64(pv.FailedOnChain))

		idx.MetricsVecMap[PollVoteMetricName].
			With(prometheus.Labels{
				VoteStatusLabel: "NotFound",
				VerifierLabel:   pv.Moniker,
			}).
			Set(float64(pv.NotFound))

		idx.MetricsVecMap[PollVoteMetricName].
			With(prometheus.Labels{
				VoteStatusLabel: "SucceededOnChain",
				VerifierLabel:   pv.Moniker,
			}).
			Set(float64(pv.SucceededOnChain))
	}
}
