package indexer

import (
	"time"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	totalMissCounterMetricName  = "bls_signature_missed_total"
	latestMissCounterMetricName = "validator_bls_signature_missed_counter"

	epochLabel = "epoch"
)

func (idx *CheckpointIndexer) initLabelsAndMetrics() {
	// indexer default to check sync well
	idx.MetricsMap[common.IndexPointerEpochMetricName] = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerEpochMetricName,
		ConstLabels: idx.PackageLabels,
	})
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName] = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockTimestampMetricName,
		ConstLabels: idx.PackageLabels,
	})

	// validator miss metrics
	idx.MetricsVecMap[totalMissCounterMetricName] = idx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        totalMissCounterMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "Total BLS signatures missed by a CometBFT Validator",
	}, []string{
		common.MonikerLabel,
		common.StatusLabel,
	})
	idx.MetricsVecMap[latestMissCounterMetricName] = idx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        latestMissCounterMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "Total amount of Validators that missed a BLS signature on the latest epoch",
	}, []string{
		common.MonikerLabel,
		common.StatusLabel,
		epochLabel,
	})
}

func (idx *CheckpointIndexer) updatePrometheusMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	idx.MetricsMap[common.IndexPointerEpochMetricName].Set(float64(indexPointer))
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName].Set((float64(indexPointerTimestamp.Unix())))
	idx.Debugf("update prometheus metrics %d epoch", indexPointer)

	// for _, ve := range bveList {
	// 	idx.MetricsVecMap[latestMissCounterMetricName].
	// 		With(prometheus.Labels{common.MonikerLabel: ve., common.StatusLabel: tmtypes.BlockIDFlag_name[model.UnknownCount]}).
	// 		Set(float64(model.UnknownCount))
	// }

	modelList, err := idx.repo.SelectTotalMissList(idx.ChainID)
	if err != nil {
		idx.Errorf("failed to update recent miss counter metric: %s", err)
	}
	for _, model := range modelList {
		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: tmtypes.BlockIDFlag_name[model.UnknownCount]}).
			Set(float64(model.UnknownCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: tmtypes.BlockIDFlag_name[model.AbsentCount]}).
			Set(float64(model.AbsentCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: tmtypes.BlockIDFlag_name[model.CommitCount]}).
			Set(float64(model.CommitCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: tmtypes.BlockIDFlag_name[model.NilCount]}).
			Set(float64(model.NilCount))
	}
}
