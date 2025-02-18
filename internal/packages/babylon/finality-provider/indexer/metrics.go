package indexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

func (idx *FinalityProviderIndexer) initLabelsAndMetrics() {
	indexPointerBlockHeightMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockHeightMetricName,
		ConstLabels: idx.PackageLabels,
	})

	latestBlockHeightMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.LatestBlockHeightMetricName,
		ConstLabels: idx.PackageLabels,
	})

	indexPointerBlockHeightMetric.Set(0)
	idx.MetricsMap[common.IndexPointerBlockHeightMetricName] = indexPointerBlockHeightMetric

	latestBlockHeightMetric.Set(0)
	idx.MetricsMap[common.LatestBlockHeightMetricName] = latestBlockHeightMetric
}

func (idx *FinalityProviderIndexer) updatePrometheusMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	idx.MetricsMap[common.IndexPointerBlockHeightMetricName].Set(float64(indexPointer))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}
