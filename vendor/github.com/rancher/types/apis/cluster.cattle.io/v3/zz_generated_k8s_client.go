package v3

import (
	"context"
	"sync"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/objectclient/dynamic"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/rest"
)

type (
	contextKeyType        struct{}
	contextClientsKeyType struct{}
)

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	ClusterAuthTokensGetter
	ClusterUserAttributesGetter
}

type Clients struct {
	Interface Interface

	ClusterAuthToken     ClusterAuthTokenClient
	ClusterUserAttribute ClusterUserAttributeClient
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	clusterAuthTokenControllers     map[string]ClusterAuthTokenController
	clusterUserAttributeControllers map[string]ClusterUserAttributeController
}

func Factory(ctx context.Context, config rest.Config) (context.Context, controller.Starter, error) {
	c, err := NewForConfig(config)
	if err != nil {
		return ctx, nil, err
	}

	cs := NewClientsFromInterface(c)

	ctx = context.WithValue(ctx, contextKeyType{}, c)
	ctx = context.WithValue(ctx, contextClientsKeyType{}, cs)
	return ctx, c, nil
}

func ClientsFrom(ctx context.Context) *Clients {
	return ctx.Value(contextClientsKeyType{}).(*Clients)
}

func From(ctx context.Context) Interface {
	return ctx.Value(contextKeyType{}).(Interface)
}

func NewClients(config rest.Config) (*Clients, error) {
	iface, err := NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewClientsFromInterface(iface), nil
}

func NewClientsFromInterface(iface Interface) *Clients {
	return &Clients{
		Interface: iface,

		ClusterAuthToken: &clusterAuthTokenClient2{
			iface: iface.ClusterAuthTokens(""),
		},
		ClusterUserAttribute: &clusterUserAttributeClient2{
			iface: iface.ClusterUserAttributes(""),
		},
	}
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = dynamic.NegotiatedSerializer
	}

	restClient, err := restwatch.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient: restClient,

		clusterAuthTokenControllers:     map[string]ClusterAuthTokenController{},
		clusterUserAttributeControllers: map[string]ClusterUserAttributeController{},
	}, nil
}

func (c *Client) RESTClient() rest.Interface {
	return c.restClient
}

func (c *Client) Sync(ctx context.Context) error {
	return controller.Sync(ctx, c.starters...)
}

func (c *Client) Start(ctx context.Context, threadiness int) error {
	return controller.Start(ctx, threadiness, c.starters...)
}

type ClusterAuthTokensGetter interface {
	ClusterAuthTokens(namespace string) ClusterAuthTokenInterface
}

func (c *Client) ClusterAuthTokens(namespace string) ClusterAuthTokenInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ClusterAuthTokenResource, ClusterAuthTokenGroupVersionKind, clusterAuthTokenFactory{})
	return &clusterAuthTokenClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ClusterUserAttributesGetter interface {
	ClusterUserAttributes(namespace string) ClusterUserAttributeInterface
}

func (c *Client) ClusterUserAttributes(namespace string) ClusterUserAttributeInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ClusterUserAttributeResource, ClusterUserAttributeGroupVersionKind, clusterUserAttributeFactory{})
	return &clusterUserAttributeClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
