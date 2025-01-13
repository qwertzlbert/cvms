package indexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

func (idx *CheckpointIndexer) initLabelsAndMetrics() {
	epochMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerEpochMetricName,
		ConstLabels: idx.PackageLabels,
	})
	idx.MetricsMap[common.IndexPointerEpochMetricName] = epochMetric

	timestampMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockTimestampMetricName,
		ConstLabels: idx.PackageLabels,
	})
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName] = timestampMetric

	// last epoch miss counting metrics
	// lastMissCounterMetric := idx.Factory.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace:   common.Namespace,
	// 	Subsystem:   subsystem,
	// 	Name:        common.RecentMissCounterMetricName,
	// 	ConstLabels: idx.PackageLabels,
	// }, []string{
	// 	common.MonikerLabel,
	// })
}

// func (idx *CheckpointIndexer) updateLastMissCounterMetric() {
// rvvList, err := vidx.repo.SelectRecentMissValidatorVoteList(vidx.ChainID)
// if err != nil {
// 	vidx.Errorf("failed to update recent miss counter metric: %s", err)
// }

// for _, rvv := range rvvList {
// 	vidx.MetricsVecMap[common.RecentMissCounterMetricName].
// 		With(prometheus.Labels{common.MonikerLabel: rvv.Moniker}).
// 		Set(float64(rvv.MissedCount))
// }
// }

func (idx *CheckpointIndexer) updatePrometheusMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	idx.MetricsMap[common.IndexPointerEpochMetricName].Set(float64(indexPointer))
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName].Set((float64(indexPointerTimestamp.Unix())))
	// idx.Debugf("update prometheus metrics %d epoch", indexPointer)
}
