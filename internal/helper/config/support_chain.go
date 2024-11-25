package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type SupportChains struct {
	Chains map[string]ChainDetail `yaml:",inline"`
}

type ChainDetail struct {
	ChainName    string   `yaml:"chain_name"`
	ProtocolType string   `yaml:"protocol_type"`
	Mainnet      bool     `yaml:"mainnet"`
	Consumer     bool     `yaml:"consumer"`
	Packages     []string `yaml:"packages"`
	SupportAsset Asset    `yaml:"support_asset"`
}

type Asset struct {
	Denom   string `yaml:"denom"`
	Decimal int    `yaml:"decimal"`
}

func (sc *SupportChains) Marshal() ([]byte, error) {
	yamlData, err := yaml.Marshal(sc)
	if err != nil {
		return nil, err
	}
	return yamlData, nil
}

func GetSupportChainConfig() (*SupportChains, error) {
	dataBytes, err := os.ReadFile(MustGetSupportChainPath("support_chains.yaml"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file")
	}

	scCfg := &SupportChains{}
	err = yaml.Unmarshal(dataBytes, scCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode config file")
	}

	if fileExists(MustGetSupportChainPath("custom_chains.yaml")) {
		ctDataBytes, err := os.ReadFile(MustGetSupportChainPath("custom_chains.yaml"))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read config file")
		}
		ctCfg := &SupportChains{}
		err = yaml.Unmarshal(ctDataBytes, ctCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode second config file")
		}
		// Merge the two configurations
		for chainName, chainDetail := range ctCfg.Chains {
			// NOTE: this will override chain config for users to use custom_chains.yaml
			if _, exists := scCfg.Chains[chainName]; exists {
				scCfg.Chains[chainName] = chainDetail
			}

			// Add custom chains by custom_chains.yaml
			scCfg.Chains[chainName] = chainDetail
		}
	}

	return scCfg, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
