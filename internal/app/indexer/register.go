package indexer

import (
	"strconv"
	"strings"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var PackageFilter string

func register(m common.Mode, f promauto.Factory, l *logrus.Logger, idb *common.IndexerDB, mc *config.MonitoringConfig, sc *config.SupportChains) error {
	l.Infof("supported packages for indexer application: %v", common.IndexPackages)
	for _, cc := range mc.ChainConfigs {
		chain := sc.Chains[cc.ChainID]
		mainnet := chain.Mainnet
		chainID := cc.ChainID
		chainName := chain.ChainName
		packages := chain.Packages
		protocolType := chain.ProtocolType
		isConsumer := chain.Consumer
		monikers := mc.Monikers

		if cc.Monikers != nil {
			l.Debugf("found individual moniker list: %v for chain: %v", cc.Monikers, chain.ChainName)
			monikers = cc.Monikers
		}

		for _, pkg := range packages {
			if ok := helper.Contains(common.IndexPackages, pkg); ok {
				if PackageFilter == "" {
					err := selectPackage(m, f, l, idb, mainnet, chainID, chainName, pkg, protocolType, isConsumer, cc, monikers)
					if err != nil {
						l.WithField("package", pkg).WithField("chain", chainName).WithField("chain_id", chainID).
							Errorf("this package was failed to start while initiating, so that the package will be skipped: %s", err)

						common.Skip.With(prometheus.Labels{
							common.ChainLabel:   chainName,
							common.ChainIDLabel: chainID,
							common.PackageLabel: pkg,
							common.MainnetLabel: strconv.FormatBool(mainnet),
							common.ErrLabel:     err.Error(),
						}).Inc()
					}
				} else if strings.Contains(string(pkg), PackageFilter) {
					l.Debugf("the package is filterd by package filter flag, only %s package is going to be register in indexer application", PackageFilter)
					err := selectPackage(m, f, l, idb, mainnet, chainID, chainName, pkg, protocolType, isConsumer, cc, monikers)
					if err != nil {
						l.WithField("package", pkg).WithField("chain", chainName).Errorf("this package is skipped by %s", err)
						common.Skip.With(prometheus.Labels{
							common.ChainLabel:   chainName,
							common.ChainIDLabel: chainID,
							common.PackageLabel: pkg,
							common.MainnetLabel: strconv.FormatBool(mainnet),
							common.ErrLabel:     err.Error(),
						}).Inc()
					}
				}
			}
		}
	}
	return nil
}
