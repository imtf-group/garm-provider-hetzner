package client

import (
	"context"
	"fmt"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/imtf-group/garm-provider-hetzner/internal/spec"
	"strconv"
)

type HcloudClient struct {
	cfg *config.Config
	api ClientInterface
}

func DeserializeInstance(instance *hcloud.Server) params.ProviderInstance {
	providerInstance := params.ProviderInstance{
		ProviderID: strconv.FormatInt(instance.ID, 10),
	}

	for key, value := range instance.Labels {
		switch key {
		case "Name":
			providerInstance.Name = value
		case "OSType":
			providerInstance.OSType = params.OSType(value)
		case "OSArch":
			providerInstance.OSArch = params.OSArch(value)
		}
	}

	switch instance.Status {
	case hcloud.ServerStatusInitializing,
		hcloud.ServerStatusRunning,
		hcloud.ServerStatusStarting,
		hcloud.ServerStatusMigrating,
		hcloud.ServerStatusRebuilding,
		hcloud.ServerStatusStopping:

		providerInstance.Status = params.InstanceRunning
	case hcloud.ServerStatusOff,
		hcloud.ServerStatusDeleting:

		providerInstance.Status = params.InstanceStopped
	default:
		providerInstance.Status = params.InstanceStatusUnknown
	}

	return providerInstance
}

func NewClient(ctx context.Context, cfg *config.Config) (*HcloudClient, error) {
	client := hcloud.NewClient(hcloud.WithToken(cfg.Token))

	hcloudClient := &HcloudClient{
		cfg: cfg,
		api: &HCloudAPI{client: client},
	}

	return hcloudClient, nil
}

func (c *HcloudClient) Config() *config.Config {
	return c.cfg
}

func (c *HcloudClient) SetConfig(cfg *config.Config) {
	c.cfg = cfg
}

func (c *HcloudClient) Api() ClientInterface {
	return c.api
}

func (c *HcloudClient) SetApi(api ClientInterface) {
	c.api = api
}

func (c *HcloudClient) CreateInstance(ctx context.Context, spec *spec.RunnerSpec) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("invalid nil runner spec")
	}

	udata, err := spec.ComposeUserData()
	if err != nil {
		return "", fmt.Errorf("failed to compose user data: %w", err)
	}

	serverType := &hcloud.ServerType{Name: spec.BootstrapParams.Flavor}
	location := &hcloud.Location{Name: spec.Location}
	image := &hcloud.Image{Name: spec.BootstrapParams.Image}

	var sshKeys []*hcloud.SSHKey
	var networks []*hcloud.Network
	var firewalls []*hcloud.ServerCreateFirewall
	var datacenter *hcloud.Datacenter
	var placementGroup *hcloud.PlacementGroup

	for _, sshKey := range spec.SSHKeys {
		sshKeys = append(sshKeys, &hcloud.SSHKey{ID: sshKey})
	}

	for _, network := range spec.Networks {
		networks = append(networks, &hcloud.Network{ID: network})
	}

	for _, firewall := range spec.Firewalls {
		firewalls = append(firewalls, &hcloud.ServerCreateFirewall{
			Firewall: hcloud.Firewall{ID: firewall},
		})
	}
	if spec.Datacenter != "" {
		datacenter = &hcloud.Datacenter{Name: spec.Datacenter}
		location = nil
	}

	if spec.PlacementGroup != 0 {
		placementGroup = &hcloud.PlacementGroup{ID: spec.PlacementGroup}
	}

	result, _, err := c.api.CreateServer(ctx, hcloud.ServerCreateOpts{
		UserData:         udata,
		Name:             spec.BootstrapParams.Name,
		StartAfterCreate: hcloud.Ptr(true),
		ServerType:       serverType,
		Image:            image,
		Location:         location,
		SSHKeys:          sshKeys,
		Networks:         networks,
		PublicNet: &hcloud.ServerCreatePublicNet{
			EnableIPv4: (!spec.DisableIPv4),
			EnableIPv6: (!spec.DisableIPv6),
		},
		Firewalls:      firewalls,
		PlacementGroup: placementGroup,
		Datacenter:     datacenter,
		Labels: map[string]string{
			"Name":               spec.BootstrapParams.Name,
			"GARM_POOL_ID":       spec.BootstrapParams.PoolID,
			"OSType":             string(spec.BootstrapParams.OSType),
			"OSArch":             string(spec.BootstrapParams.OSArch),
			"GARM_CONTROLLER_ID": spec.ControllerID,
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}
	return strconv.FormatInt(result.Server.ID, 10), nil
}

func (c *HcloudClient) DeleteInstance(ctx context.Context, instance string) error {
	server, err := c.GetInstance(ctx, instance)
	if err != nil {
		return err
	}
	_, err = c.api.DeleteServer(ctx, server)
	if err != nil {
		return fmt.Errorf("error during deletion: %v (ID: %d)", err, server.ID)
	}
	return nil
}

func (c *HcloudClient) GetInstance(ctx context.Context, instance string) (*hcloud.Server, error) {
	server, _, err := c.api.GetServer(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving the serverID: %v", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server with ID %q not found", instance)
	}
	return server, nil
}

func (c *HcloudClient) GetAllInstances(ctx context.Context) ([]*hcloud.Server, error) {
	servers, err := c.api.GetAllServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return servers, nil
}

func (c *HcloudClient) StartInstance(ctx context.Context, instance string) error {
	server, err := c.GetInstance(ctx, instance)
	if err != nil {
		return err
	}
	if server.Status != hcloud.ServerStatusOff {
		return fmt.Errorf("instance %s cannot be started in %s state", instance, server.Status)
	}
	_, _, err = c.api.StartServer(ctx, server)
	if err != nil {
		return fmt.Errorf("error while starting: %v (ID: %d)", err, server.ID)
	}
	return nil
}

func (c *HcloudClient) StopInstance(ctx context.Context, instance string) error {
	server, err := c.GetInstance(ctx, instance)
	if err != nil {
		return err
	}
	if server.Status != hcloud.ServerStatusRunning && server.Status != hcloud.ServerStatusStarting {
		return fmt.Errorf("instance %s cannot be stopped in %s state", instance, server.Status)
	}
	_, _, err = c.api.StopServer(ctx, server)
	if err != nil {
		return fmt.Errorf("error while stopping: %v (ID: %d)", err, server.ID)
	}
	return nil
}
