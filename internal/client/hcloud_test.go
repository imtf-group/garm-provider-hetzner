package client

import (
	"context"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/imtf-group/garm-provider-hetzner/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestDeserializeInstance(t *testing.T) {
	server := &hcloud.Server{
		ID:     123456,
		Status: hcloud.ServerStatusRunning,
		Labels: map[string]string{
			"Name":         "garm-0000",
			"GARM_POOL_ID": "09876-54321",
			"OSType":       "linux",
			"OSArch":       "amd64",
		},
	}
	instance := DeserializeInstance(server)
	expected := params.ProviderInstance{
		ProviderID: "123456",
		Name:       "garm-0000",
		Status:     params.InstanceRunning,
		OSType:     "linux",
		OSArch:     "amd64",
	}
	assert.Equal(t, instance, expected)
}

func TestGetApi(t *testing.T) {
	mockAPI := new(MockHCloudAPI)
	client := &HcloudClient{api: mockAPI}
	assert.Equal(t, client.api, client.Api())
}

func TestSetApi(t *testing.T) {
	mockAPI := new(MockHCloudAPI)
	client := &HcloudClient{}
	client.SetApi(mockAPI)
	assert.Equal(t, client.api, mockAPI)
}

func TestGetConfig(t *testing.T) {
	mockAPI := new(MockHCloudAPI)
	client := &HcloudClient{
		api: mockAPI,
		cfg: &config.Config{
			Location: "location",
			Token:    "token",
		},
	}
	assert.Equal(t, client.cfg, client.Config())
}

func TestSetConfig(t *testing.T) {
	client := &HcloudClient{}
	cfg := &config.Config{
		Location: "location",
		Token:    "token",
	}
	client.SetConfig(cfg)
	assert.Equal(t, client.cfg, cfg)
}

func TestCreateInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	Mocktools := params.RunnerApplicationDownload{
		OS:           hcloud.Ptr("linux"),
		Architecture: hcloud.Ptr("amd64"),
		DownloadURL:  hcloud.Ptr("MockURL"),
		Filename:     hcloud.Ptr("garm-runner"),
	}
	client := &HcloudClient{api: mockAPI}

	spec := &spec.RunnerSpec{
		Datacenter: "fsn1-dc14",
		Location:   "fsn1",
		BootstrapParams: params.BootstrapInstance{
			Name:   "test-runner",
			PoolID: "pool-1",
			OSType: "linux",
			Flavor: "cx22",
			OSArch: "amd64",
		},
		ControllerID:   "controller-xyz",
		PlacementGroup: 111111,
		Networks:       []int64{22222, 33333},
		Firewalls:      []int64{44444},
		Tools:          Mocktools,
		DisableIPv6:    true,
	}

	mockAPI.On("CreateServer", mock.Anything, mock.MatchedBy(func(opts hcloud.ServerCreateOpts) bool {
		assert.Nil(t, opts.Location)
		assert.NotNil(t, opts.Datacenter)
		assert.NotNil(t, opts.PlacementGroup, 111111)
		assert.Equal(t, opts.Networks, []*hcloud.Network{
			&hcloud.Network{ID: 22222},
			&hcloud.Network{ID: 33333},
		})
		assert.Equal(t, opts.Firewalls, []*hcloud.ServerCreateFirewall{
			&hcloud.ServerCreateFirewall{
				Firewall: hcloud.Firewall{ID: 44444},
			},
		})
		assert.Equal(t, opts.PublicNet, &hcloud.ServerCreatePublicNet{
			EnableIPv4: true,
			EnableIPv6: false,
		})
		assert.Equal(t, "fsn1-dc14", opts.Datacenter.Name)
		return true
	})).Return(hcloud.ServerCreateResult{Server: &hcloud.Server{ID: 123456}}, &hcloud.Response{}, nil)

	serverID, err := client.CreateInstance(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, serverID, "123456")
	mockAPI.AssertExpectations(t)
}

func TestDeleteInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456}, &hcloud.Response{}, nil)

	mockAPI.On("DeleteServer", mock.Anything, mock.MatchedBy(func(server *hcloud.Server) bool {
		assert.Equal(t, server.ID, int64(123456))
		return true
	})).Return(&hcloud.Response{}, nil)

	err := client.DeleteInstance(context.Background(), "123456")
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestGetInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Name: "my-server"}, &hcloud.Response{}, nil)

	server, err := client.GetInstance(context.Background(), "123456")
	assert.NoError(t, err)
	assert.Equal(t, server.ID, int64(123456))
	mockAPI.AssertExpectations(t)
}

func TestGetAllInstances(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetAllServers", mock.Anything).Return([]*hcloud.Server{
		&hcloud.Server{
			ID:   123456,
			Name: "my-server",
		},
	}, nil)

	servers, err := client.GetAllInstances(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, servers[0].ID, int64(123456))
	mockAPI.AssertExpectations(t)
}

func TestStartInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusOff}, &hcloud.Response{}, nil)

	mockAPI.On("StartServer", mock.Anything, mock.MatchedBy(func(server *hcloud.Server) bool {
		assert.Equal(t, server.ID, int64(123456))
		return true
	})).Return(&hcloud.Action{}, &hcloud.Response{}, nil)

	err := client.StartInstance(context.Background(), "123456")
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestStartInstanceAlreadyStarted(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusRunning}, &hcloud.Response{}, nil)

	err := client.StartInstance(context.Background(), "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be started in running state")
	mockAPI.AssertExpectations(t)
}

func TestStopInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusRunning}, &hcloud.Response{}, nil)

	mockAPI.On("StopServer", mock.Anything, mock.MatchedBy(func(server *hcloud.Server) bool {
		assert.Equal(t, server.ID, int64(123456))
		return true
	})).Return(&hcloud.Action{}, &hcloud.Response{}, nil)

	err := client.StopInstance(context.Background(), "123456")
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestStopInstanceAlreadyStopped(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServer", mock.Anything, mock.MatchedBy(func(instance string) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusOff}, &hcloud.Response{}, nil)

	err := client.StopInstance(context.Background(), "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be stopped in off state")
	mockAPI.AssertExpectations(t)
}
