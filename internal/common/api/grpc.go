package api

import (
	"context"
	"fmt"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	grpchelper "github.com/cosmostation/cvms/internal/helper/grpc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func GetValidatorsbyGRPC(c common.CommonClient) ([]types.CosmosValidator, error) {

	// init context
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	totalValidators := make([]types.CosmosValidator, 0)

	// create grpc client connection
	grpcConnection, err := grpc.NewClient(c.GetGRPCEndPoint(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		c.Errorf("grpc request error: %s", err.Error())
		return nil, common.ErrFailedCreateGrpcConnection
	}
	defer grpcConnection.Close()

	var headerMD metadata.MD
	reflectionClient := grpcreflect.NewClientAuto(
		metadata.NewOutgoingContext(ctx, headerMD),
		grpcConnection,
	)
	offset := 0
	totalValidatorsCount := 0
	var validators []types.CosmosValidator
	for ok := true; ok; ok = (len(totalValidators) < totalValidatorsCount) {
		// get on-chain validators
		resp, err := grpchelper.GrpcDynamicQuery(
			ctx,                                    // common context
			reflectionClient,                       // grpc reflection client
			grpcConnection,                         // grpc connection stub
			types.CommonValidatorConsGrpcQueryPath, // grpc query method
			fmt.Sprintf(`{"pagination" :{"offset": "%d"}}`, offset), // grpc query payload
		)

		if err != nil {
			return nil, errors.Errorf("grpc request err: %s", err.Error())
		}

		// json unmarsharling received validators data
		validators, totalVC, err := parser.CosmosValidatorParserGCP([]byte(resp))
		if err != nil {
			c.Errorf("parser error: %s", err)
			return nil, errors.Errorf("got data, but failed to parse the data, error %s", err.Error())
		}
		totalValidatorsCount = int(totalVC)
		totalValidators = append(totalValidators, validators...)
		offset = len(totalValidators)
	}

	if len(totalValidators) != totalValidatorsCount {
		c.Warningf("number of validators %d doesn't match the total validators count: %d", len(validators), totalValidatorsCount)
	}
	c.Infof("found: %d cosmos validators in %d pages", len(totalValidators), totalValidatorsCount)

	return totalValidators, nil
}
