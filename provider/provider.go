package provider

import (
	"context"
	"fmt"
	execution "github.com/cloudbase/garm-provider-common/execution/v0.1.0"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/imtf-group/garm-provider-hetzner/internal/client"
	"github.com/imtf-group/garm-provider-hetzner/internal/spec"
)

var Version = "v0.0.1"

func NewHcloudProvider(ctx context.Context, configPath, controllerID string) (execution.ExternalProvider, error) {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	client, err := client.NewClient(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("error getting the client: %w", err)
	}
	return &HcloudProvider{
		controllerID: controllerID,
		client:       client,
	}, nil
}

type HcloudProvider struct {
	controllerID string
	client       *client.HcloudClient
}

func (a *HcloudProvider) CreateInstance(ctx context.Context, bootstrapParams params.BootstrapInstance) (params.ProviderInstance, error) {
	spec, err := spec.GetRunnerSpecFromBootstrapParams(a.client.Config(), bootstrapParams, a.controllerID)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to get runner spec: %w", err)
	}

	instanceID, err := a.client.CreateInstance(ctx, spec)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to create instance: %w", err)
	}

	instance := params.ProviderInstance{
		ProviderID: instanceID,
		Name:       spec.BootstrapParams.Name,
		OSType:     spec.BootstrapParams.OSType,
		OSArch:     spec.BootstrapParams.OSArch,
		Status:     "running",
	}

	return instance, nil
}

func (a *HcloudProvider) DeleteInstance(ctx context.Context, instance string) error {
	if err := a.client.DeleteInstance(ctx, instance); err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

func (a *HcloudProvider) GetInstance(ctx context.Context, instance string) (params.ProviderInstance, error) {
	server, err := a.client.GetInstance(ctx, instance, false)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to get VM details: %w", err)
	}
	if server == nil {
		return params.ProviderInstance{}, nil
	}

	providerInstance := client.DeserializeInstance(server)

	return providerInstance, nil
}

func (a *HcloudProvider) ListInstances(ctx context.Context, poolID string) ([]params.ProviderInstance, error) {
	servers, err := a.client.GetAllInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	var providerInstances []params.ProviderInstance
	for _, server := range servers {
		for key, value := range server.Labels {
			if key == "GARM_POOL_ID" && value == poolID {
				providerInstances = append(providerInstances, client.DeserializeInstance(server))
				break
			}
		}
	}
	return providerInstances, nil
}

func (a *HcloudProvider) RemoveAllInstances(ctx context.Context) error {
	return nil
}

func (a *HcloudProvider) Stop(ctx context.Context, instance string, force bool) error {
	return a.client.StopInstance(ctx, instance)
}

func (a *HcloudProvider) Start(ctx context.Context, instance string) error {
	return a.client.StartInstance(ctx, instance)
}

func (a *HcloudProvider) GetVersion(ctx context.Context) string {
	return Version
}
