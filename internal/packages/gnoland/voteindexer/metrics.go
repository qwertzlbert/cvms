package gnolandvoteindexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	RecentMissCounterMetricName = "recent_miss_counter"
)

func (vidx *VoteIndexer) initLabelsAndMetrics() {
	indexPointerBlockHeightMetric := vidx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockHeightMetricName,
		ConstLabels: vidx.PackageLabels,
	})
	indexPointerBlockTimestampMetric := vidx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockTimestampMetricName,
		ConstLabels: vidx.PackageLabels,
	})
	indexPointerBlockHeightMetric.Set(0)
	vidx.MetricsMap[common.IndexPointerBlockHeightMetricName] = indexPointerBlockHeightMetric
	indexPointerBlockTimestampMetric.Set(0)
	vidx.MetricsMap[common.IndexPointerBlockTimestampMetricName] = indexPointerBlockTimestampMetric

	recentMissCounterMetric := vidx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        RecentMissCounterMetricName,
		ConstLabels: vidx.PackageLabels,
	}, []string{
		common.MonikerLabel,
	})

	vidx.MetricsVecMap[RecentMissCounterMetricName] = recentMissCounterMetric
}

func (vidx *VoteIndexer) updateRecentMissCounterMetric() {
	rvvList, err := vidx.repo.SelectRecentMissValidatorVoteList(vidx.ChainID)
	if err != nil {
		vidx.Errorf("failed to update recent miss counter metric: %s", err)
	}

	for _, rvv := range rvvList {
		vidx.MetricsVecMap[RecentMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: rvv.Moniker}).
			Set(float64(rvv.MissedCount))
	}
}

func (idx *VoteIndexer) updateRootMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	idx.MetricsMap[common.IndexPointerBlockHeightMetricName].Set(float64(indexPointer))
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName].Set(float64(indexPointerTimestamp.Unix()))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}
