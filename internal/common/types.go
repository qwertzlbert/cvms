package common

import (
	"context"
	"io"
	"net/http"

	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	GRPCClient *resty.Client
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
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
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
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Wrapf(err, "api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
	}
	return resp.Body(), nil
}

func NewCommonApp(p Packager) CommonApp {
	restyLogger := logrus.New()
	restyLogger.Out = io.Discard
	rpcClient := NewRestyClient().SetLogger(restyLogger)
	apiClient := NewRestyClient().SetLogger(restyLogger)
	grpcClient := resty.New().
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryMaxWaitTimeDuration).
		SetRetryMaxWaitTime(retryMaxWaitTimeDuration).
		SetLogger(restyLogger)
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

func (c *CommonClient) SetGRPCEndPoint(endpoint string) *resty.Client {
	return c.GRPCClient.SetBaseURL(endpoint)
}

func (c *CommonClient) GetGRPCEndPoint() string {
	return c.GRPCClient.BaseURL
}

func NewOptionalClient(entry *logrus.Entry) CommonClient {
	restyLogger := logrus.New()
	restyLogger.Out = io.Discard
	rpcClient := NewRestyClient().SetLogger(restyLogger)
	apiClient := NewRestyClient().SetLogger(restyLogger)
	return CommonClient{rpcClient, apiClient, nil, entry}
}
