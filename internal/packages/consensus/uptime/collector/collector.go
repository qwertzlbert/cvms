package collector

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/router"
	"github.com/cosmostation/cvms/internal/packages/consensus/uptime/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ common.CollectorStart = Start
	_ common.CollectorLoop  = loop
)

const (
	Subsystem                         = "uptime"
	SubsystemSleep                    = 10 * time.Second
	UnHealthSleep                     = 10 * time.Second
	MissBlockCounterMetricName        = "missed_blocks_counter"
	JailedMetricName                  = "jailed"
	StakedTokensMetricName            = "staked_tokens_total"
	SignedBlocksWindowMetricName      = "signed_blocks_window"
	MinSignedPerWindowMetricName      = "min_signed_per_window"
	DowntimeJailDurationMetricName    = "downtime_jail_duration"
	SlashFractionDowntimeMetricName   = "slash_fraction_downtime"
	SlashFractionDoubleSignMetricName = "slash_fraction_double_sign"
	bondedValidatorsTotalMetricName   = "bonded_validators_total"
	activeValidatorsTotalMetricName   = "active_validators_total"
	minSeatPriceMetric                = "min_seat_price"
	validatorCommissionMetric         = "validator_commission_rate"
)

func Start(p common.Packager) error {
	if ok := helper.Contains(types.SupportedProtocolTypes, p.ProtocolType); ok {
		exporter := common.NewExporter(p)
		for _, rpc := range p.RPCs {
			exporter.SetRPCEndPoint(rpc)
			break
		}
		for _, api := range p.APIs {
			exporter.SetAPIEndPoint(api)
			break
		}

		if p.IsConsumerChain {
			exporter.OptionalClient = common.NewOptionalClient(exporter.Entry)
			for _, rpc := range p.ProviderEndPoints.RPCs {
				exporter.OptionalClient.SetRPCEndPoint(rpc)
				break
			}
			for _, api := range p.ProviderEndPoints.APIs {
				exporter.OptionalClient.SetAPIEndPoint(api)
				break
			}
		}
		go loop(exporter, p)
		return nil
	}
	return errors.Errorf("unsupported protocol type: %s", p.ProtocolType)
}

func loop(exporter *common.Exporter, p common.Packager) {
	rootLabels := common.BuildRootLabels(p)
	packageLabels := common.BuildPackageLabels(p)

	// metrics for each validator
	uptimeMetric := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        MissBlockCounterMetricName,
		ConstLabels: packageLabels,
	}, []string{
		common.MonikerLabel,
		common.ValidatorAddressLabel,
		common.ConsensusAddressLabel,
		common.ProposerAddressLabel,
	})

	jailedMetric := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        JailedMetricName,
		ConstLabels: packageLabels,
	}, []string{
		common.MonikerLabel,
		common.ValidatorAddressLabel,
		common.ConsensusAddressLabel,
		common.ProposerAddressLabel,
	})

	stakedTokensMetric := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        StakedTokensMetricName,
		Help:        "The total staked tokens of a validator",
		ConstLabels: packageLabels,
	}, []string{
		common.MonikerLabel,
		common.ValidatorAddressLabel,
		common.ConsensusAddressLabel,
		common.ProposerAddressLabel,
	})

	validatorCommissionMetric := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        validatorCommissionMetric,
		Help:        "The commission rate of a validator",
		ConstLabels: packageLabels,
	}, []string{
		common.MonikerLabel,
		common.ValidatorAddressLabel,
		common.ConsensusAddressLabel,
		common.ProposerAddressLabel,
	})

	// metrics for each chain
	signedBlocksWindowMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        SignedBlocksWindowMetricName,
		ConstLabels: packageLabels,
	})
	minSignedPerWindowMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        MinSignedPerWindowMetricName,
		ConstLabels: packageLabels,
	})
	downtimeJailDurationMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        DowntimeJailDurationMetricName,
		Help:        "The duration a node will be jailed for downtime (in seconds)",
		ConstLabels: packageLabels,
	})
	slashFractionDowntimeMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        SlashFractionDowntimeMetricName,
		Help:        "The fraction of validator's stake slashed for downtime",
		ConstLabels: packageLabels,
	})
	slashFractionDoubleSignMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        SlashFractionDoubleSignMetricName,
		Help:        "The fraction of validator's stake slashed for double signing",
		ConstLabels: packageLabels,
	})
	bondedValidatorsTotalMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        bondedValidatorsTotalMetricName,
		Help:        "The total number of bonded validators",
		ConstLabels: packageLabels,
	})
	activeValidatorsTotalMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        activeValidatorsTotalMetricName,
		Help:        "The total number of active validators",
		ConstLabels: packageLabels,
	})
	minSeatPriceMetric := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        minSeatPriceMetric,
		Help:        "The minimum amount of stake required to get a seat as an active validator",
		ConstLabels: packageLabels,
	})

	isUnhealth := false
	for {
		// node health check
		if isUnhealth {
			healthAPIs := healthcheck.FilterHealthEndpoints(p.APIs, p.ProtocolType)
			for _, api := range healthAPIs {
				exporter.SetAPIEndPoint(api)
				exporter.Infoln("client endpoint will be changed with health endpoint for this package")
				isUnhealth = false
				break
			}
			healthRPCs := healthcheck.FilterHealthRPCEndpoints(p.RPCs, p.ProtocolType)
			for _, rpc := range healthRPCs {
				exporter.SetRPCEndPoint(rpc)
				exporter.Warnf("RPC endpoint will be changed with health endpoint for this package: %s", rpc)
				isUnhealth = false
				break
			}
			if len(healthAPIs) == 0 || len(healthRPCs) == 0 {
				isUnhealth = true
				exporter.Errorln("failed to get any health endpoints from healthcheck filter, retry sleep 10s")
				time.Sleep(UnHealthSleep)
				continue
			}
		}

		status, err := router.GetStatus(exporter, p)
		if err != nil {
			common.Health.With(rootLabels).Set(0)
			common.Ops.With(rootLabels).Inc()
			isUnhealth = true

			exporter.Errorf("failed to update metrics err: %s and going to sleep %s...", err, SubsystemSleep.String())
			time.Sleep(SubsystemSleep)

			continue
		}

		if p.Mode == common.NETWORK {
			// update metrics by each validators
			for _, item := range status.Validators {
				uptimeMetric.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
						common.ProposerAddressLabel:  item.ProposerAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(float64(item.MissedBlockCounter))
				jailedMetric.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
						common.ProposerAddressLabel:  item.ProposerAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(float64(item.IsTomstoned))

				stakedTokensMetric.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
						common.ProposerAddressLabel:  item.ProposerAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(item.StakedTokens)

				validatorCommissionMetric.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
						common.ProposerAddressLabel:  item.ProposerAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(item.CommissionRate)

			}
		} else {
			// update metrics by each validators
			for _, item := range status.Validators {
				if ok := helper.Contains(p.Monikers, item.Moniker); ok {
					uptimeMetric.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
							common.ProposerAddressLabel:  item.ProposerAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(float64(item.MissedBlockCounter))
					jailedMetric.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
							common.ProposerAddressLabel:  item.ProposerAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(float64(item.IsTomstoned))
					stakedTokensMetric.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
							common.ProposerAddressLabel:  item.ProposerAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(item.StakedTokens)
					validatorCommissionMetric.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.ConsensusAddressLabel: item.ValidatorConsensusAddress,
							common.ProposerAddressLabel:  item.ProposerAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(item.CommissionRate)
				}
			}
		}

		// update metrics by each chain
		signedBlocksWindowMetric.Set(status.SignedBlocksWindow)
		minSignedPerWindowMetric.Set(status.MinSignedPerWindow)
		downtimeJailDurationMetric.Set(status.DowntimeJailDuration)
		slashFractionDowntimeMetric.Set(status.SlashFractionDowntime)
		slashFractionDoubleSignMetric.Set(status.SlashFractionDoubleSign)
		bondedValidatorsTotalMetric.Set(float64(status.BondedValidatorsTotal))
		activeValidatorsTotalMetric.Set(float64(len(status.Validators)))
		minSeatPriceMetric.Set(float64(status.MinimumSeatPrice))

		exporter.Infof("updated metrics successfully and going to sleep %s ...", SubsystemSleep.String())

		// update health and ops
		common.Health.With(rootLabels).Set(1)
		common.Ops.With(rootLabels).Inc()

		// sleep
		time.Sleep(SubsystemSleep)
	}
}
