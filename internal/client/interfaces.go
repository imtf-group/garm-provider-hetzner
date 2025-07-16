package client

import (
	"context"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/stretchr/testify/mock"
)

type ClientInterface interface {
	GetAllServers(ctx context.Context) ([]*hcloud.Server, error)
	GetServer(ctx context.Context, instance string) (*hcloud.Server, *hcloud.Response, error)
	CreateServer(ctx context.Context, opts hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, *hcloud.Response, error)
	DeleteServer(ctx context.Context, server *hcloud.Server) (*hcloud.Response, error)
	StartServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error)
	StopServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error)
}

type HCloudAPI struct {
	client *hcloud.Client
}

func (r *HCloudAPI) GetAllServers(ctx context.Context) ([]*hcloud.Server, error) {
	return r.client.Server.All(ctx)
}

func (r *HCloudAPI) GetServer(ctx context.Context, instance string) (*hcloud.Server, *hcloud.Response, error) {
	return r.client.Server.Get(ctx, instance)
}

func (r *HCloudAPI) DeleteServer(ctx context.Context, server *hcloud.Server) (*hcloud.Response, error) {
	_, resp, err := r.client.Server.DeleteWithResult(ctx, server)
	return resp, err
}

func (r *HCloudAPI) CreateServer(ctx context.Context, opts hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, *hcloud.Response, error) {
	return r.client.Server.Create(ctx, opts)
}

func (r *HCloudAPI) StartServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {
	return r.client.Server.Poweron(ctx, server)
}

func (r *HCloudAPI) StopServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {
	return r.client.Server.Poweroff(ctx, server)
}

type MockHCloudAPI struct {
	mock.Mock
}

func (m *MockHCloudAPI) GetAllServers(ctx context.Context) ([]*hcloud.Server, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*hcloud.Server), args.Error(1)
}

func (m *MockHCloudAPI) GetServer(ctx context.Context, instance string) (*hcloud.Server, *hcloud.Response, error) {
	args := m.Called(ctx, instance)
	var srv *hcloud.Server
	if tmp := args.Get(0); tmp != nil {
		srv = tmp.(*hcloud.Server)
	}
	return srv, args.Get(1).(*hcloud.Response), args.Error(2)
}

func (m *MockHCloudAPI) CreateServer(ctx context.Context, opts hcloud.ServerCreateOpts) (hcloud.ServerCreateResult, *hcloud.Response, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(hcloud.ServerCreateResult), args.Get(1).(*hcloud.Response), args.Error(2)
}

func (m *MockHCloudAPI) DeleteServer(ctx context.Context, server *hcloud.Server) (*hcloud.Response, error) {
	args := m.Called(ctx, server)
	return args.Get(0).(*hcloud.Response), args.Error(1)
}

func (m *MockHCloudAPI) StartServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {
	args := m.Called(ctx, server)
	return args.Get(0).(*hcloud.Action), args.Get(1).(*hcloud.Response), args.Error(2)
}

func (m *MockHCloudAPI) StopServer(ctx context.Context, server *hcloud.Server) (*hcloud.Action, *hcloud.Response, error) {
	args := m.Called(ctx, server)
	return args.Get(0).(*hcloud.Action), args.Get(1).(*hcloud.Response), args.Error(2)
}
