# Consumer Chain Metrics

Consumer chains (e.g. Stride or Neutron) as defined by cosmos [Interchain Security Standard](https://tutorials.cosmos.network/academy/2-cosmos-concepts/14-interchain-security.html) 
rely on a provider chain (e.g. cosmoshub) for 
providing security. This means if the validator node of a consumer chain misbehaves, the linked
validator address on the provider chain will be penalized instead. 

CVMS supports tracking metrics of consumer chain validators by resolving the linked 
provider validator via calls to the [Interchain_Security Module](https://buf.build/cosmos/interchain-security).

## Example: Getting Validator metrics for stride

To correctly track the metrics of a consumer chain the `provider_nodes` variable needs to be defined!

```yaml
monikers:
  - 'Cosmostation'

chains:

  - display_name: 'stride'
    chain_id: 'stride-1'
    nodes:
      - rpc: 'https://rpc-stride.endpoint.xyz'
        api: 'https://lcd-stride.endpoint.xyz'
        grpc: 'grpc-stride.endpoint.xyz:9090'

    provider_nodes:

      - rpc: 'https://rpc-cosmos.endpoint.xyz'
        api: 'https://lcd-cosmos.endpoint.xyz'
        grpc: 'grpc-cosmos.endpoint.xyz:9090'

```

The metrics gathered by such a config are tagged by using the addresses of the linked provider
validator like visible below:

```
# HELP cvms_uptime_jailed 
# TYPE cvms_uptime_jailed gauge
cvms_uptime_jailed{chain="stride",chain_id="stride-1",mainnet="true",moniker="SuperAwesomeValidator",package="uptime",proposer_address="ABCDEFG...",table_chain_id="stride_1",validator_consensus_address="cosmosvalconsxyz",validator_operator_address="cosmosvaloperxyz"} 0
```

## How  it works

To return the correct validator related metrics CVMS will do the following steps:

1. Lookup the `consumer chain ID` on the provider network
    - For this the [interchain_security provider module](https://buf.build/cosmos/interchain-security/docs/main:interchain_security.ccv.provider.v1#interchain_security.ccv.provider.v1.QueryConsumerChainsRequest) API is used
    - This returns a list of consumer chains supported by the provider network
2. Lookup provider validators securing consumer chain validators
    - again the [interchain_security provder module](https://buf.build/cosmos/interchain-security/docs/main:interchain_security.ccv.provider.v1#interchain_security.ccv.provider.v1.QueryConsumerValidatorsRequest) is called on the provider network
    - This returns a list of validators securing validators on the consumer chain
3. Get the Bech32 consensus address of the consumer validator
    - To calculate the related bech32 address on the consumer chain the Human Readable Part (HRP)
    of the consumer chain aswell as the validators pubkey on the provider chain is required
    - Using this information the consensus address (`valcons`) can be calculated from the pubkey
4. Get the `signing_infos` statistics from the consumer chain
    - Using the calculated consensus address the signer statistics can be looked up by querying the
    signing infos API of the [Slashing](https://buf.build/cosmos/cosmos-sdk/docs/main:cosmos.slashing.v1beta1#cosmos.slashing.v1beta1.QuerySigningInfoRequest) Module