package common

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/go-resty/resty/v2"

	"github.com/jhump/protoreflect/grpcreflect"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"

	grpchelper "github.com/cosmostation/cvms/internal/helper/grpc"
)

// TODO: All Methods in VoteIndexer, we need to add here?
type IIndexer interface {
	Start() error
	Loop(lastIndexPointerHeight int64)
	FetchValidatorInfoList() error
}

// TODO
type ICollector interface {
	Start(p Packager) error
}

// Collector Function Sig
type CollectorStart func(Packager) error
type CollectorLoop func(*Exporter, Packager)

// Client
type ClientType int

const (
	RPC ClientType = iota
	API
	GRPC
)

// Methods
type Method int

const (
	GET Method = iota
	POST
)

// Application Mode

type Mode int

const (
	INVALID_APP Mode = -1   // Invalid Case
	NETWORK     Mode = iota // Network Mode to provide network status overview
	VALIDATOR               // Validator Mode to provide whole chains' status overview about validator
)

func (a Mode) String() string {
	switch {
	case a == NETWORK:
		return "Validator Monitoring System"
	case a == VALIDATOR:
		return "White List"
	default:
		return "Invalid Mode"
	}
}

type CommonClient struct {
	RPCClient  Client
	APIClient  Client
	GRPCClient Client
	*logrus.Entry
}
type CommonApp struct {
	CommonClient
	EndPoint string
	// optional client
	OptionalClient CommonClient
}

type Client interface {
	SetEndpoint(endpoint string) Client
	GetEndpoint() (string, error)
	Get(context context.Context, uri string) ([]byte, error)
	Post(context context.Context, uri string, body ...[]byte) ([]byte, error)
}

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

type RestyClient struct {
	client   *resty.Client
	endpoint string
	logger   *logrus.Logger
}

func (rc *RestyClient) SetLogger(logger *logrus.Logger) Client {
	rc.logger = logger
	rc.client.SetLogger(logger)
	return rc
}

func (rc *RestyClient) SetEndpoint(endpoint string) Client {
	rc.endpoint = endpoint
	rc.client.SetBaseURL(endpoint)

	return rc
}

func (rc *RestyClient) GetEndpoint() (string, error) {
	if rc.endpoint == "" {
		return "", errors.New("endpoint is not set")
	}
	return rc.endpoint, nil
}

func NewRestyClient() *RestyClient {
	rc := &RestyClient{}
	rc.client = resty.New().
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryMaxWaitTimeDuration).
		SetRetryMaxWaitTime(retryMaxWaitTimeDuration).
		SetHeader("Content-Type", "application/json")

	return rc
}

func (rc *RestyClient) Get(context context.Context, uri string) ([]byte, error) {
	resp, err := rc.client.R().
		SetContext(context).
		Get(rc.endpoint + uri)
	if err != nil {
		rc.logger.Debugf("GET request failed: %s", err.Error())
		return nil, err
	}
	rc.logger.Debugf("Received %d status code from %s", resp.StatusCode(), resp.Request.URL)
	if resp.StatusCode() != http.StatusOK {
		rc.logger.Errorf("GET request failed with status code: %d", resp.StatusCode())
		return nil, errors.Wrapf(err, "api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}
	return resp.Body(), nil
}

func (rc *RestyClient) Post(context context.Context, uri string, body ...[]byte) ([]byte, error) {
	resp, err := rc.client.R().
		SetHeader("Content-Type", "application/json").
		SetContext(context).
		SetBody(body).
		Post(rc.endpoint + uri)
	if err != nil {
		rc.logger.Debugf("POST request failed: %s", err.Error())
		return nil, err
	}
	rc.logger.Debugf("Received %d status code from %s", resp.StatusCode(), resp.Request.URL)
	// This ensures that the response status code is within the 200-202 range,
	// which indicates a successful POST request.
	if resp.StatusCode() <= http.StatusOK && resp.StatusCode() <= http.StatusAccepted {
		rc.logger.Errorf("POST request failed with status code: %d", resp.StatusCode())
		return nil, errors.Wrapf(err, "api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}
	return resp.Body(), nil
}

func NewCommonApp(p Packager) CommonApp {
	rpcClient := NewRestyClient().SetLogger(p.Logger)
	apiClient := NewRestyClient().SetLogger(p.Logger)
	grpcClient := NewGrpcClient().SetLogger(p.Logger)
	entry := p.Logger.WithFields(
		logrus.Fields{
			logger.FieldKeyChain:   p.ChainName,
			logger.FieldKeyChainID: p.ChainID,
			logger.FieldKeyPackage: p.Package,
		})
	commonClient := CommonClient{rpcClient, apiClient, grpcClient, entry}
	return CommonApp{
		commonClient,
		"",
		CommonClient{},
	}
}

func (c *CommonClient) SetRPCEndPoint(endpoint string) Client {
	return c.RPCClient.SetEndpoint(endpoint)
}

func (c *CommonClient) GetRPCEndPoint() string {
	endpoint, _ := c.RPCClient.GetEndpoint()
	return endpoint
}

func (c *CommonClient) SetAPIEndPoint(endpoint string) Client {
	return c.APIClient.SetEndpoint(endpoint)
}

func (c *CommonClient) GetAPIEndPoint() string {
	endpoint, _ := c.APIClient.GetEndpoint()
	return endpoint
}

func (c *CommonClient) SetGRPCEndPoint(endpoint string) Client {
	return c.GRPCClient.SetEndpoint(endpoint)
}

func (c *CommonClient) GetGRPCEndPoint() string {
	endpoint, _ := c.GRPCClient.GetEndpoint()
	return endpoint
}

func NewOptionalClient(entry *logrus.Entry) CommonClient {
	rpcClient := NewRestyClient().SetLogger(entry.Logger)
	apiClient := NewRestyClient().SetLogger(entry.Logger)
	grpcClient := NewGrpcClient().SetLogger(entry.Logger)
	return CommonClient{rpcClient, apiClient, grpcClient, entry}
}
