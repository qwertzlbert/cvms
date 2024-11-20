package collector

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/router"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ common.CollectorStart = Start
	_ common.CollectorLoop  = loop
)

// NOTE: this is for solo mode
var packageMonikers []string

const (
	Subsystem      = "yoda"
	SubsystemSleep = 60 * time.Second
	UnHealthSleep  = 10 * time.Second

	YodaStatusMetricName = "status"
	// This metric is used to track the total number of request misses,
	// from the startup to the application.
	YodaTotalRequestMisses = "total_miss_counter"
	// Collects the maximum number of current misses from all requests the validator
	// has to respond and which are not yet expired
	YodaMaxMisses = "max_miss_counter"
	// Collects the miss counter per request. Note: probably makes sense to limit to the last 100 or so requests.
	YodaMissesPerRequest = "miss_counter"
	// Shows the maximum number of blocks a yoda oracle has time to respond to a request
	YodaRequestSlashWindow = "slash_window"
)

func Start(p common.Packager) error {
	if ok := helper.Contains(types.SupportedChains, p.ChainName); ok {
		packageMonikers = p.Monikers
		for _, api := range p.APIs {
			client := common.NewExporter(p)
			client.SetAPIEndPoint(api)
			go loop(client, p)
			break
		}
		return nil
	}
	return errors.Errorf("unsupported chain: %s", p.ChainName)
}

func loop(c *common.Exporter, p common.Packager) {
	rootLabels := common.BuildRootLabels(p)
	packageLabels := common.BuildPackageLabels(p)

	// each validators
	yodaStatusMetrics := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaStatusMetricName,
	}, []string{
		common.ValidatorAddressLabel,
		common.MonikerLabel,
	})

	// yodaTotalMisses := p.Factory.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace:   common.Namespace,
	// 	Subsystem:   Subsystem,
	// 	ConstLabels: packageLabels,
	// 	Name:        YodaTotalRequestMisses,
	// }, []string{
	// 	common.ValidatorAddressLabel,
	// 	common.MonikerLabel,
	// })

	// yodaMaxMisses := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace:   common.Namespace,
	// 	Subsystem:   Subsystem,
	// 	ConstLabels: packageLabels,
	// 	Name:        YodaMaxMisses,
	// }, []string{
	// 	common.ValidatorAddressLabel,
	// 	common.MonikerLabel,
	// })

	// each validator and non-expired request
	// yodaRequestMisses := p.Factory.NewCounterVec(prometheus.CounterOpts{
	// 	Namespace:   common.Namespace,
	// 	Subsystem:   Subsystem,
	// 	ConstLabels: packageLabels,
	// 	Name:        YodaMissesPerRequest,
	// }, []string{
	// 	common.ValidatorAddressLabel,
	// 	common.MonikerLabel,
	// 	common.YodaRequestIDLabel,
	// })

	// each chain
	yodaSlashWindow := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaRequestSlashWindow,
	})

	isUnhealth := false
	for {
		// node health check
		if isUnhealth {
			healthEndpoints := healthcheck.FilterHealthEndpoints(p.APIs, p.ProtocolType)
			for _, endpoint := range healthEndpoints {
				c.SetAPIEndPoint(endpoint)
				c.Infoln("client endpoint will be changed with health endpoint for this package")
				isUnhealth = false
				break
			}
			if len(healthEndpoints) == 0 {
				c.Errorln("failed to get any health endpoints from healthcheck filter, retry sleep 10s")
				time.Sleep(UnHealthSleep)
				continue
			}
		}

		// start collect status
		status, err := router.GetStatus(c, p.ChainName)
		if err != nil {
			common.Health.With(rootLabels).Set(0)
			common.Ops.With(rootLabels).Inc()
			isUnhealth = true

			c.Logger.Errorf("failed to update metrics: %s", err.Error())
			time.Sleep(SubsystemSleep)

			continue
		}

		if p.Mode == common.NETWORK {
			// update metrics for each validators
			for _, item := range status.Validators {
				yodaStatusMetrics.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(item.IsActive)
			}

		} else {
			// filter metrics for only specific validator
			for _, item := range status.Validators {
				if ok := helper.Contains(packageMonikers, item.Moniker); ok {
					yodaStatusMetrics.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(float64(item.IsActive))
				}
			}
		}

		// update metrics for each chain
		yodaSlashWindow.Set(status.SlashWindow)

		c.Infof("updated %s metrics successfully and going to sleep %s ...", Subsystem, SubsystemSleep.String())

		// update health and ops
		common.Health.With(rootLabels).Set(1)
		common.Ops.With(rootLabels).Inc()

		// sleep
		time.Sleep(SubsystemSleep)
	}
}
