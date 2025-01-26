package client

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"

	grpchelper "github.com/cosmostation/cvms/internal/helper/grpc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GrpcClient struct {
	client   *grpc.ClientConn
	endpoint string
	logger   *logrus.Logger
}

func (gc *GrpcClient) SetLogger(logger *logrus.Logger) Client {
	gc.logger = logger
	return gc
}

// Basically returns a new GrpcClient instance as it does not
// support changing the endpoint
func (gc *GrpcClient) SetEndpoint(endpoint string) Client {

	var dialOptions []grpc.DialOption

	tlsConf := &tls.Config{
		NextProtos: []string{"h2"}, // only allow HTTP/2
		MinVersion: tls.VersionTLS12,
	}
	gCreds := credentials.NewTLS(tlsConf)

	// Simple handshake check to determine if the server supports TLS
	_, err := tls.Dial("tcp", endpoint, tlsConf)
	if err != nil {
		var recordHeaderError tls.RecordHeaderError
		if errors.As(err, &recordHeaderError) {
			gc.logger.Warnf("TLS handshake failed")
			dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

		} else {
			gc.logger.Errorf("error setting up gRPC connection: %s", err.Error())
		}
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(gCreds))
	}

	client, err := grpc.NewClient(endpoint, dialOptions...)
	if err != nil {
		gc.logger.Errorf("error setting up grpc client: %s", err.Error())
		return nil
	}
	gc.client = client
	gc.endpoint = endpoint
	return gc
}

func (gc *GrpcClient) GetEndpoint() (string, error) {
	if gc.endpoint == "" {
		gc.logger.Error("grpc endpoint is not set")
		return "", errors.New("endpoint is not set")
	}
	return gc.endpoint, nil
}

// As Get is similar to Post with an empty body, it is implemented as a simple wrapper around Post
func (gc *GrpcClient) Get(ctx context.Context, uri string) ([]byte, error) {
	respJSON, err := gc.Post(ctx, uri, []byte(``))
	if err != nil {
		gc.logger.Errorf("grpc api failed to get: %s", err.Error())
		return nil, err
	}
	return respJSON, nil
}

func (gc *GrpcClient) Post(ctx context.Context, uri string, body ...[]byte) ([]byte, error) {

	var protoResolver grpchelper.CosmosAnyMessageResolver
	// var reflectionHeader metadata.MD
	var respHeader metadata.MD

	if body == nil {
		gc.logger.Errorf("grpc api requires a body for POST request. Use Get method instead.")
		return nil, errors.New("grpc api requires a body for POST request")
	}

	// wakes up the client if it's not ready
	if gc.client.GetState() != connectivity.Ready {
		gc.logger.Info("grpc client is not ready yet. Waking up...")
		gc.client.Connect()
	}

	refCtx := metadata.NewOutgoingContext(ctx, make(metadata.MD))

	reflectionClient := grpcreflect.NewClientAuto(
		refCtx,
		gc.client,
	)

	descriptor, err := grpchelper.ResolveMessage(uri, reflectionClient)
	if err != nil {
		gc.logger.Errorf("grpc api failed to resolve proto message: %s", err.Error())
		return nil, err
	}

	msg, err := grpchelper.CreateMessage(descriptor, string(body[0]))
	if err != nil {
		gc.logger.Errorf("grpc api failed to create proto message: %s", err.Error())
		return nil, err
	}

	// This only exists to support the current variable syntax
	// e.g. "cosmos.staking.v1beta1.Query.Validators" to actually use the method it
	// must be re-formatted to "/cosmos.staking.v1beta1.Query/Validators"
	response := dynamicpb.NewMessage(descriptor.Output())
	idx := strings.LastIndex(uri, ".")
	fullMethodName := "/" + uri[:idx] + "/" + uri[idx+1:]

	peer := &peer.Peer{}
	gc.logger.Debugf("Invoking method: %s", fullMethodName)
	err = gc.client.Invoke(ctx, fullMethodName, msg, response, grpc.Header(&respHeader), grpc.Peer(peer))
	gc.logger.Debugf(
		"GRPC connection info: Protocol: %s Target: %s Local Address: %s Authentication Type: %s",
		peer.Addr.Network(),
		peer.Addr.String(),
		peer.LocalAddr.String(),
		peer.AuthInfo.AuthType(),
	)
	gc.logger.Debugf("GRPC header returned: %+v", respHeader)
	if err != nil {
		gc.logger.Errorf("failed to invoke grpc request: %s", err.Error())
		return nil, err
	}
	marshaller := protojson.MarshalOptions{
		AllowPartial:    true,
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: false,
		Resolver:        protoResolver,
	}

	respJSON, err := marshaller.Marshal(response)
	if err != nil {
		gc.logger.Errorf("grpc api failed to marshal string with grpc data: %s", err.Error())
		return nil, err
	}
	return respJSON, nil
}

func NewGrpcClient() *GrpcClient {
	gc := &GrpcClient{}
	return gc
}
