package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	grpchelper "github.com/cosmostation/cvms/internal/helper/grpc"
	"github.com/cosmostation/cvms/internal/packages/duty/eventnonce/types"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func GetEventNonceStatusByGRPC(
	c *common.Exporter,
	commonOrchestratorPath string, commonOrchestratorParser func([]byte) (string, error),
	commonEventNonceQueryPath string, commonEventNonceParser func([]byte) (float64, error),
) (types.CommonEventNonceStatus, error) {
	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	creds := insecure.NewCredentials()

	// check the port is 443 or not
	if strings.Contains(c.GetGRPCEndPoint(), ":443") {
		tlsConf := &tls.Config{
			NextProtos: []string{"h2"}, // only allow HTTP/2
			MinVersion: tls.VersionTLS12,
		}

		creds = credentials.NewTLS(tlsConf)
	}

	// create grpc connection
	grpcConnection, err := grpc.NewClient(c.GetGRPCEndPoint(),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		c.Errorf("grpc request error: %s", err.Error())
		return types.CommonEventNonceStatus{}, common.ErrFailedCreateGrpcConnection
	}
	defer grpcConnection.Close()

	// create grpc reflection client & stub
	stub := grpcdynamic.NewStub(grpcConnection)

	var headerMD metadata.MD
	reflectionClient := grpcreflect.NewClientV1Alpha(
		metadata.NewOutgoingContext(ctx, headerMD),
		reflectpb.NewServerReflectionClient(grpcConnection),
	)

	// get on-chain validators
	resp, err := grpchelper.GrpcDynamicQuery(
		ctx,                                  // common context
		reflectionClient,                     // grpc reflection client
		&stub,                                // grpc query stub
		types.CommonValidatorGrpcQueryPath,   // grpc query method
		types.CommonValidatorGrpcQueryOption, // grpc query payload
	)

	if err != nil {
		c.Errorf("grpc request err: %s", err.Error())
		return types.CommonEventNonceStatus{}, common.ErrFailedGrpcRequest
	}

	// json unmarsharling received validators data
	var validators types.CommonValidatorsQueryResponse
	if err := json.Unmarshal([]byte(resp), &validators); err != nil {
		c.Errorf("parser error: %s", err)
		return types.CommonEventNonceStatus{}, common.ErrFailedJsonUnmarshal
	}

	// init channel and waitgroup
	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	var methodDescriptor *desc.MethodDescriptor
	validatorResults := make([]types.ValidatorStatus, 0)

	// add wg by the number of total validators
	wg.Add(len(validators.Validators))

	// get validators orchestrator address
	for idx, item := range validators.Validators {
		// in only first time, make a method descriptor by using grpc reflection client
		if idx == 0 {
			methodDescriptor, err = grpchelper.GrpcMakeDescriptor(
				reflectionClient,       // grpc reflection client
				commonOrchestratorPath, // grpc reflection method path
			)
			if err != nil {
				c.Errorln("grpc api err: failed to make method descprtior")
				return types.CommonEventNonceStatus{}, common.ErrFailedGrpcRequest
			}
		}

		validatorOperatorAddress := item.OperatorAddress
		validatorMoniker := item.Description.Moniker
		commonOrchestratorPayload := fmt.Sprintf(`{"validator_address":"%s"}`, validatorOperatorAddress)

		go func(ch chan helper.Result) {
			defer wg.Done()

			resp, err := grpchelper.GrpcInvokeQuery(
				ctx,                       // common context
				methodDescriptor,          // grpc query method descriptor
				&stub,                     // grpc query stub
				commonOrchestratorPayload, // grpc query payload
			)

			if err != nil {
				// NOTE: we need to modify this logic in the future for general cases

				// case.1 : not registered validators for gravity-bridge
				if strings.Contains(err.Error(), "codespace gravity code 3: invalid: No validator") {
					c.Infof("got empty orchestrator address for %s, so saved empty string", validatorOperatorAddress)
					ch <- helper.Result{Success: true, Item: types.ValidatorStatus{
						ValidatorOperatorAddress: validatorOperatorAddress,
						OrchestratorAddress:      "",
						Moniker:                  validatorMoniker,
					}}
					return
				}

				c.Errorf("grpc error: %s for %s", err, commonOrchestratorPayload)
				ch <- helper.Result{Success: false, Item: nil}
				return
			}

			orchestratorAddress, err := commonOrchestratorParser([]byte(resp))
			if err != nil {
				c.Errorf("grpc error: %v", err)
				ch <- helper.Result{Success: false, Item: nil}
				return
			}

			if orchestratorAddress == "" {
				// not registered validators
				c.Infof("got empty orchestrator address for %s, so saved empty string", validatorOperatorAddress)
				ch <- helper.Result{Success: true, Item: types.ValidatorStatus{
					ValidatorOperatorAddress: validatorOperatorAddress,
					OrchestratorAddress:      "",
					Moniker:                  validatorMoniker,
				}}
				return
			}

			ch <- helper.Result{
				Success: true,
				Item: types.ValidatorStatus{
					ValidatorOperatorAddress: validatorOperatorAddress,
					OrchestratorAddress:      orchestratorAddress,
					Moniker:                  validatorMoniker,
				},
			}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	// close channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect validator's orch
	errorCounter := 0
	for r := range ch {
		if r.Success {
			validatorResults = append(validatorResults, r.Item.(types.ValidatorStatus))
			continue
		}
		errorCounter++
	}

	if errorCounter > 0 {
		return types.CommonEventNonceStatus{}, fmt.Errorf("unexpected errors was found: total %d errors", errorCounter)
	}

	c.Debugf("total validators: %d, total orchestrator result len: %d", len(validators.Validators), len(validatorResults))

	// init channel and waitgroup for go-routine
	ch = make(chan helper.Result)
	eventNonceResults := make([]types.ValidatorStatus, 0)

	// add wg by the number of total orchestrators
	wg = sync.WaitGroup{}
	wg.Add(len(validatorResults))

	// get eventnonce by each orchestrator
	for idx, item := range validatorResults {
		if idx == 0 {
			methodDescriptor, err = grpchelper.GrpcMakeDescriptor(
				reflectionClient,          // grpc reflection client
				commonEventNonceQueryPath, // grpc reflection method path
			)
			if err != nil {
				c.Errorln("grpc api err: failed to make method descprtior")
				return types.CommonEventNonceStatus{}, common.ErrFailedGrpcRequest
			}
		}

		orchestratorAddress := item.OrchestratorAddress
		validatorOperatorAddress := item.ValidatorOperatorAddress
		validatorMoniker := item.Moniker
		payload := fmt.Sprintf(`{"address":"%s"}`, orchestratorAddress)

		go func(ch chan helper.Result) {
			defer wg.Done()
			if orchestratorAddress == "" {
				c.Warnf("skipped empty orchestrator address for %s", validatorOperatorAddress)
				ch <- helper.Result{
					Success: true,
					Item: types.ValidatorStatus{
						ValidatorOperatorAddress: validatorOperatorAddress,
						OrchestratorAddress:      orchestratorAddress,
						EventNonce:               0,
					},
				}
				return
			}

			resp, err := grpchelper.GrpcInvokeQuery(
				ctx,              // common context
				methodDescriptor, // grpc query method descriptor
				&stub,            // grpc query stub
				payload,          // grpc query payload
			)
			if err != nil {
				c.Errorf("grpc error: %s", err)
				ch <- helper.Result{Success: false, Item: nil}
				return
			}

			eventNonce, err := commonEventNonceParser([]byte(resp))
			if err != nil {
				c.Errorf("grpc error: %v", err)
				ch <- helper.Result{Success: false, Item: nil}
				return
			}

			ch <- helper.Result{
				Success: true,
				Item: types.ValidatorStatus{
					Moniker:                  validatorMoniker,
					ValidatorOperatorAddress: validatorOperatorAddress,
					OrchestratorAddress:      orchestratorAddress,
					EventNonce:               eventNonce,
				},
			}
		}(ch)
		time.Sleep(10 * time.Millisecond)
	}

	// close channels
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect results
	errorCounter = 0
	for r := range ch {
		if r.Success {
			eventNonceResults = append(eventNonceResults, r.Item.(types.ValidatorStatus))
			continue
		}
		errorCounter++
	}

	if errorCounter > 0 {
		return types.CommonEventNonceStatus{}, fmt.Errorf("unexpected errors was found: total %d errors", errorCounter)
	}

	// find heighest eventnonce in the results
	heighestEventNonce := float64(0)
	for idx, item := range eventNonceResults {
		if idx == 0 {
			heighestEventNonce = item.EventNonce
			c.Debugln("set heighest nonce:", heighestEventNonce)
		}

		if item.EventNonce > heighestEventNonce {
			c.Debugln("changed heightest nonce from: ", heighestEventNonce, "to: ", item.EventNonce)
			heighestEventNonce = item.EventNonce
		}
	}

	return types.CommonEventNonceStatus{
		HeighestNonce: heighestEventNonce,
		Validators:    eventNonceResults,
	}, nil
}
