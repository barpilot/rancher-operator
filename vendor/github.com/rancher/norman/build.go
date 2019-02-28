package norman

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/rancher/norman/api"
	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/leader"
	"github.com/rancher/norman/pkg/kwrapper/k8s"
	"github.com/rancher/norman/pkg/remotedialer"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/norman/store/proxy"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type serverContextKey struct{}

func GetServer(ctx context.Context) *Server {
	return ctx.Value(serverContextKey{}).(*Server)
}

func (c *Config) Build(ctx context.Context, opts *Options) (context.Context, *Server, error) {
	var (
		err      error
		starters []controller.Starter
	)

	if c.Name == "" {
		return ctx, nil, errors.New("name must be set on norman.Config")
	}

	if opts == nil {
		opts = &Options{
			K8sMode: "external",
		}
	}

	r := &Runtime{
		AllSchemas: types.NewSchemas(),
	}

	server := &Server{
		Config:  c,
		Runtime: r,
	}

	ctx = context.WithValue(ctx, serverContextKey{}, server)

	ctx, err = c.defaults(ctx, r, *opts)
	if err != nil {
		return ctx, nil, err
	}

	for _, schema := range c.Schemas {
		r.AllSchemas.AddSchemas(schema)
	}

	c.createCRDs(ctx, r)

	ctx, starters, err = c.clients(ctx, r)
	if err != nil {
		return ctx, nil, err
	}

	if c.CustomizeSchemas != nil {
		if err := c.CustomizeSchemas(ctx, c.ClientGetter, r.AllSchemas); err != nil {
			return ctx, nil, err
		}
	}

	if c.GlobalSetup != nil {
		ctx, err = c.GlobalSetup(ctx)
		if err != nil {
			return ctx, nil, err
		}
	}

	if err := c.registerControllers(ctx, c.PerServerControllers); err != nil {
		return ctx, nil, err
	}

	if !opts.DisableControllers {
		go c.masterControllers(ctx, starters, r)
	}

	if !c.DisableAPI {
		if err := c.apiServer(ctx, r); err != nil {
			return ctx, nil, err
		}
	}

	if c.PreStart != nil {
		if err := c.PreStart(ctx); err != nil {
			return ctx, nil, err
		}
	}

	if !opts.DisableControllers {
		err = controller.SyncThenStart(ctx, c.Threadiness, starters...)
	}

	return ctx, server, err
}

func (c *Config) apiServer(ctx context.Context, r *Runtime) error {
	if c.K3s.RemoteDialerAuthorizer != nil && r.K3sTunnelServer == nil {
		r.K3sTunnelServer = remotedialer.New(c.K3s.RemoteDialerAuthorizer, remotedialer.DefaultErrorWriter)
	}

	server := api.NewAPIServer()
	if err := server.AddSchemas(r.AllSchemas); err != nil {
		return err
	}

	r.APIHandler = server

	if c.APISetup != nil {
		if err := c.APISetup(ctx, server); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) registerControllers(ctx context.Context, controllers []ControllerRegister) error {
	for _, controller := range controllers {
		if err := controller(ctx); err != nil {
			return fmt.Errorf("failed to start controller: %v", err)
		}
	}

	return nil
}

func (c *Config) masterControllers(ctx context.Context, starters []controller.Starter, r *Runtime) {
	leader.RunOrDie(ctx, c.LeaderLockNamespace, c.Name, c.K8sClient, func(ctx context.Context) {
		var (
			err error
		)

		if c.MasterSetup != nil {
			ctx, err = c.MasterSetup(ctx)
			if err != nil {
				logrus.Fatalf("failed to setup master: %v", err)
			}
		}

		err = c.registerControllers(ctx, c.MasterControllers)
		if err != nil {
			logrus.Fatalf("failed to register master controllers: %v", err)
		}

		if err := controller.SyncThenStart(ctx, c.Threadiness, starters...); err != nil {
			logrus.Fatalf("failed to start master controllers: %v", err)
		}

		<-ctx.Done()
	})
}

func (c *Config) defaults(ctx context.Context, r *Runtime, opts Options) (context.Context, error) {
	var (
		err error
	)

	if c.Threadiness <= 0 {
		c.Threadiness = 5
	}

	if opts.KubeConfig != "" {
		c.KubeConfig = opts.KubeConfig
	}

	if c.Config == nil {
		mode := "auto"
		if opts.K8sMode != "" {
			mode = opts.K8sMode
		}

		envConfig := os.Getenv("KUBECONFIG")
		if c.KubeConfig != "" {
			envConfig = c.KubeConfig
		} else if c.IgnoredKubeConfigEnv {
			envConfig = ""
		}

		if c.K3s.DataDir != "" && c.K3s.RemoteDialerAuthorizer != nil {
			ctx, r.K3sServerConfig, r.K3sTunnelServer, err = k8s.NewK3sConfig(ctx, c.K3s.DataDir, c.K3s.RemoteDialerAuthorizer)
			if err != nil {
				return ctx, err
			}
		}

		r.Embedded, ctx, c.Config, err = k8s.GetConfig(ctx, mode, envConfig)
		if err != nil {
			return ctx, err
		}
	}

	r.LocalConfig = c.Config

	if c.ClientGetter == nil {
		cg, err := proxy.NewClientGetterFromConfig(*c.Config)
		if err != nil {
			return ctx, err
		}
		c.ClientGetter = cg
	}

	return ctx, nil
}

func (c *Config) createCRDs(ctx context.Context, r *Runtime) {
	factory := &crd.Factory{ClientGetter: c.ClientGetter}

	for version, types := range c.CRDs {
		factory.BatchCreateCRDs(ctx, c.CRDStorageContext, r.AllSchemas, version, types...)
	}

	factory.BatchWait()
}

func (c *Config) clients(ctx context.Context, r *Runtime) (context.Context, []controller.Starter, error) {
	var (
		starter  controller.Starter
		starters []controller.Starter
		err      error
	)

	if c.K8sClient == nil {
		c.K8sClient, err = kubernetes.NewForConfig(c.Config)
		if err != nil {
			return ctx, nil, err
		}
	}

	if c.APIExtClient == nil {
		c.APIExtClient, err = clientset.NewForConfig(c.Config)
		if err != nil {
			return ctx, nil, err
		}
	}

	for _, clientFactory := range c.Clients {
		ctx, starter, err = clientFactory(ctx, *c.Config)
		if err != nil {
			return ctx, nil, err
		}
		starters = append(starters, starter)
	}

	return ctx, starters, nil
}
