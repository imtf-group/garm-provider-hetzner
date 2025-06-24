// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Cloudbase Solutions SRL
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/imtf-group/garm-provider-hetzner/internal/client"
	"github.com/imtf-group/garm-provider-hetzner/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"testing"
)

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           hcloud.Ptr("linux"),
			Architecture: hcloud.Ptr("amd64"),
			DownloadURL:  hcloud.Ptr("MockURL"),
			Filename:     hcloud.Ptr("garm-runner"),
		}, nil
	}
	bootstrapParams := params.BootstrapInstance{
		Name:   "garm-instance",
		Flavor: "cx22",
		Image:  "ubuntu-22.04",
		Tools: []params.RunnerApplicationDownload{
			{
				OS:           hcloud.Ptr("linux"),
				Architecture: hcloud.Ptr("amd64"),
				DownloadURL:  hcloud.Ptr("MockURL"),
				Filename:     hcloud.Ptr("garm-runner"),
			},
		},
		OSType:     params.Linux,
		OSArch:     params.Amd64,
		PoolID:     "my-pool",
		ExtraSpecs: json.RawMessage(`{}`),
	}
	expectedInstance := params.ProviderInstance{
		ProviderID: providerID,
		Name:       "garm-instance",
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)

	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("CreateServer", ctx, mock.Anything).Return(hcloud.ServerCreateResult{
		Server: &hcloud.Server{
			ID: serverID,
		},
	}, &hcloud.Response{}, nil)
	result, err := provider.CreateInstance(ctx, bootstrapParams)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)
}

func TestCreateInstanceError(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           hcloud.Ptr("linux"),
			Architecture: hcloud.Ptr("amd64"),
			DownloadURL:  hcloud.Ptr("MockURL"),
			Filename:     hcloud.Ptr("garm-runner"),
		}, nil
	}
	bootstrapParams := params.BootstrapInstance{
		Name:   "garm-instance",
		Flavor: "cx22",
		Image:  "ubuntu-22.04",
		Tools: []params.RunnerApplicationDownload{
			{
				OS:           hcloud.Ptr("linux"),
				Architecture: hcloud.Ptr("amd64"),
				DownloadURL:  hcloud.Ptr("MockURL"),
				Filename:     hcloud.Ptr("garm-runner"),
			},
		},
		OSType:     params.Linux,
		OSArch:     params.Amd64,
		PoolID:     "my-pool",
		ExtraSpecs: json.RawMessage(`{}`),
	}
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)

	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("CreateServer", ctx, mock.Anything).Return(hcloud.ServerCreateResult{
		Server: &hcloud.Server{
			ID: serverID,
		},
	}, &hcloud.Response{}, fmt.Errorf("Error while creating instance"))
	_, err = provider.CreateInstance(ctx, bootstrapParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error while creating instance")
}

func TestDeleteInstance(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID: serverID,
	}, &hcloud.Response{}, nil)
	mockAPI.On("DeleteServer", ctx, &hcloud.Server{
		ID: serverID,
	}).Return(&hcloud.Response{}, nil)
	err = provider.DeleteInstance(ctx, providerID)
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestDeleteInstanceError(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)

	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID: serverID,
	}, &hcloud.Response{}, nil)
	mockAPI.On("DeleteServer", ctx, &hcloud.Server{
		ID: serverID,
	}).Return(&hcloud.Response{}, fmt.Errorf("Error while deleting instance"))
	err = provider.DeleteInstance(ctx, providerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error while deleting instance")
	mockAPI.AssertExpectations(t)
}

func TestGetInstance(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
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
	expectedInstance := params.ProviderInstance{
		ProviderID: "123456",
		Name:       "garm-0000",
		Status:     params.InstanceRunning,
		OSType:     "linux",
		OSArch:     "amd64",
	}
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(server, &hcloud.Response{}, nil)
	instance, err := provider.GetInstance(ctx, providerID)
	assert.NoError(t, err)
	assert.Equal(t, instance, expectedInstance)
	mockAPI.AssertExpectations(t)
}

func TestGetInstanceError(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
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
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(
		server, &hcloud.Response{},
		fmt.Errorf("Error retrieving instance"))
	_, err = provider.GetInstance(ctx, providerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error retrieving instance")
	mockAPI.AssertExpectations(t)
}

func TestListInstances(t *testing.T) {
	ctx := context.Background()
	servers := []*hcloud.Server{
		&hcloud.Server{
			ID:     123456,
			Status: hcloud.ServerStatusRunning,
			Labels: map[string]string{
				"Name":         "garm-0000",
				"GARM_POOL_ID": "09876-54321",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
		&hcloud.Server{
			ID:     234567,
			Status: hcloud.ServerStatusOff,
			Labels: map[string]string{
				"Name":         "garm-0001",
				"GARM_POOL_ID": "09876-54321",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
		&hcloud.Server{
			ID:     234567,
			Status: hcloud.ServerStatusOff,
			Labels: map[string]string{
				"Name":         "garm-0002",
				"GARM_POOL_ID": "09876-54322",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
	}
	expectedInstances := []params.ProviderInstance{
		params.ProviderInstance{
			ProviderID: "123456",
			Name:       "garm-0000",
			Status:     params.InstanceRunning,
			OSType:     "linux",
			OSArch:     "amd64",
		},
		params.ProviderInstance{
			ProviderID: "234567",
			Name:       "garm-0001",
			Status:     params.InstanceStopped,
			OSType:     "linux",
			OSArch:     "amd64",
		},
	}
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	mockAPI.On("GetAllServers", ctx).Return(servers, nil)
	instances, err := provider.ListInstances(ctx, "09876-54321")
	assert.NoError(t, err)
	assert.Equal(t, instances, expectedInstances)
	mockAPI.AssertExpectations(t)
}

func TestListInstancesError(t *testing.T) {
	ctx := context.Background()
	servers := []*hcloud.Server{
		&hcloud.Server{
			ID:     123456,
			Status: hcloud.ServerStatusRunning,
			Labels: map[string]string{
				"Name":         "garm-0000",
				"GARM_POOL_ID": "09876-54321",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
		&hcloud.Server{
			ID:     234567,
			Status: hcloud.ServerStatusOff,
			Labels: map[string]string{
				"Name":         "garm-0001",
				"GARM_POOL_ID": "09876-54321",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
		&hcloud.Server{
			ID:     234567,
			Status: hcloud.ServerStatusOff,
			Labels: map[string]string{
				"Name":         "garm-0002",
				"GARM_POOL_ID": "09876-54322",
				"OSType":       "linux",
				"OSArch":       "amd64",
			},
		},
	}
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	mockAPI.On("GetAllServers", ctx).Return(servers, fmt.Errorf("error while listing instances"))
	_, err := provider.ListInstances(ctx, "09876-54321")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while listing instances")
	mockAPI.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusRunning,
	}, &hcloud.Response{}, nil)
	mockAPI.On("StopServer", ctx, &hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusRunning,
	}).Return(&hcloud.Action{}, &hcloud.Response{}, nil)
	err = provider.Stop(ctx, providerID, true)
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestStopError(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusRunning,
	}, &hcloud.Response{}, nil)
	mockAPI.On("StopServer", ctx, &hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusRunning,
	}).Return(&hcloud.Action{}, &hcloud.Response{}, fmt.Errorf("error while stopping"))
	err = provider.Stop(ctx, providerID, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while stopping")
	mockAPI.AssertExpectations(t)
}

func TestStopStopped(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusOff,
	}, &hcloud.Response{}, nil)
	err = provider.Stop(ctx, providerID, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be stopped in off state")
	mockAPI.AssertExpectations(t)
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusOff,
	}, &hcloud.Response{}, nil)
	mockAPI.On("StartServer", ctx, &hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusOff,
	}).Return(&hcloud.Action{}, &hcloud.Response{}, nil)
	err = provider.Start(ctx, providerID)
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestStartError(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusOff,
	}, &hcloud.Response{}, nil)
	mockAPI.On("StartServer", ctx, &hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusOff,
	}).Return(&hcloud.Action{}, &hcloud.Response{}, fmt.Errorf("error while starting"))
	err = provider.Start(ctx, providerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error while starting")
	mockAPI.AssertExpectations(t)
}

func TestStartRunning(t *testing.T) {
	ctx := context.Background()
	providerID := "123456"
	mockAPI := new(client.MockHCloudAPI)
	provider := &HcloudProvider{
		controllerID: "controllerID",
		client:       &client.HcloudClient{},
	}
	config := &config.Config{
		Location: "nbg1",
		Token:    "mysecret",
	}
	provider.client.SetConfig(config)
	provider.client.SetApi(mockAPI)
	serverID, err := strconv.ParseInt(providerID, 10, 64)
	assert.NoError(t, err)
	mockAPI.On("GetServerByID", ctx, serverID).Return(&hcloud.Server{
		ID:     serverID,
		Status: hcloud.ServerStatusRunning,
	}, &hcloud.Response{}, nil)
	err = provider.Start(ctx, providerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be started in running state")
	mockAPI.AssertExpectations(t)
}

// func TestStop(t *testing.T) {
// 	ctx := context.Background()
// 	instanceID := "i-1234567890abcdef0"
// 	provider := &AwsProvider{
// 		controllerID: "controllerID",
// 		awsCli:       &client.AwsCli{},
// 	}
// 	config := &config.Config{
// 		Region:   "us-east-1",
// 		SubnetID: "subnet-123456",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockComputeClient := new(client.MockComputeClient)
// 	provider.awsCli.SetConfig(config)
// 	provider.awsCli.SetClient(mockComputeClient)

// 	mockComputeClient.On("StopInstances", ctx, &ec2.StopInstancesInput{
// 		InstanceIds: []string{instanceID},
// 	}, mock.Anything).Return(&ec2.StopInstancesOutput{}, nil)
// 	err := provider.Stop(ctx, instanceID, false)
// 	assert.NoError(t, err)
// }

// func TestStartStoppedInstance(t *testing.T) {
// 	ctx := context.Background()
// 	instanceID := "i-1234567890abcdef0"
// 	provider := &AwsProvider{
// 		controllerID: "controllerID",
// 		awsCli:       &client.AwsCli{},
// 	}
// 	config := &config.Config{
// 		Region:   "us-east-1",
// 		SubnetID: "subnet-123456",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockComputeClient := new(client.MockComputeClient)
// 	provider.awsCli.SetConfig(config)
// 	provider.awsCli.SetClient(mockComputeClient)

// 	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
// 		InstanceIds: []string{instanceID},
// 		Filters: []types.Filter{
// 			{
// 				Name:   aws.String("instance-state-name"),
// 				Values: []string{"pending", "running", "stopping", "stopped"},
// 			},
// 		},
// 	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: aws.String(instanceID),
// 						State: &types.InstanceState{
// 							Name: types.InstanceStateNameStopped,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}, nil)
// 	mockComputeClient.On("StartInstances", ctx, &ec2.StartInstancesInput{
// 		InstanceIds: []string{instanceID},
// 	}, mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)
// 	err := provider.Start(ctx, instanceID)
// 	assert.NoError(t, err)
// }

// func TestStartStoppingInstance(t *testing.T) {
// 	ctx := context.Background()
// 	instanceID := "i-1234567890abcdef0"
// 	provider := &AwsProvider{
// 		controllerID: "controllerID",
// 		awsCli:       &client.AwsCli{},
// 	}
// 	config := &config.Config{
// 		Region:   "us-east-1",
// 		SubnetID: "subnet-123456",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockComputeClient := new(client.MockComputeClient)
// 	provider.awsCli.SetConfig(config)
// 	provider.awsCli.SetClient(mockComputeClient)

// 	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
// 		InstanceIds: []string{instanceID},
// 		Filters: []types.Filter{
// 			{
// 				Name:   aws.String("instance-state-name"),
// 				Values: []string{"pending", "running", "stopping", "stopped"},
// 			},
// 		},
// 	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: aws.String(instanceID),
// 						State: &types.InstanceState{
// 							Name: types.InstanceStateNameStopping,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}, nil)
// 	mockComputeClient.On("StartInstances", ctx, &ec2.StartInstancesInput{
// 		InstanceIds: []string{instanceID},
// 	}, mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)
// 	err := provider.Start(ctx, instanceID)
// 	assert.Error(t, err)
// 	assert.Equal(t, "instance "+instanceID+" cannot be started in stopping state", err.Error())
// }
