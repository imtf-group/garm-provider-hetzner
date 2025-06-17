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
		OS:           hcloud.String("linux"),
		Architecture: hcloud.String("amd64"),
		DownloadURL:  hcloud.String("MockURL"),
		Filename:     hcloud.String("garm-runner"),
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

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
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

func TestDeleteInstanceInvalidValue(t *testing.T) {
	client := &HcloudClient{}
	err := client.DeleteInstance(context.Background(), "not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestGetInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
		assert.Equal(t, serverID, int64(123456))
		return true
	})).Return(&hcloud.Server{ID: 123456, Name: "my-server"}, &hcloud.Response{}, nil)

	server, err := client.GetInstance(context.Background(), "123456")
	assert.NoError(t, err)
	assert.Equal(t, server.ID, int64(123456))
	mockAPI.AssertExpectations(t)
}

func TestGetInstanceInvalidValue(t *testing.T) {
	client := &HcloudClient{}
	server, err := client.GetInstance(context.Background(), "not-a-number")
	assert.Error(t, err)
	assert.Nil(t, server)
	assert.Contains(t, err.Error(), "invalid syntax")
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

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
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

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusRunning}, &hcloud.Response{}, nil)

	err := client.StartInstance(context.Background(), "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be started in running state")
	mockAPI.AssertExpectations(t)
}

func TestStartInstanceInvalidValue(t *testing.T) {
	client := &HcloudClient{}
	err := client.StartInstance(context.Background(), "not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestStopInstance(t *testing.T) {
	mockAPI := new(MockHCloudAPI)

	client := &HcloudClient{api: mockAPI}

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
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

	mockAPI.On("GetServerByID", mock.Anything, mock.MatchedBy(func(serverID int64) bool {
		return true
	})).Return(&hcloud.Server{ID: 123456, Status: hcloud.ServerStatusOff}, &hcloud.Response{}, nil)

	err := client.StopInstance(context.Background(), "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be stopped in off state")
	mockAPI.AssertExpectations(t)
}

func TestStopInstanceInvalidValue(t *testing.T) {
	client := &HcloudClient{}
	err := client.StopInstance(context.Background(), "not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
}

// 	ctx := context.Background()
// 	id, err := client.CreateInstance(ctx, spec)
// 	require.NoError(t, err)
// 	require.Equal(t, "123456", id)
// }

// func TestStartInstance(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	instanceId := "i-1234567890abcdef0"
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	mockClient.On("StartInstances", ctx, mock.MatchedBy(func(input *ec2.StartInstancesInput) bool {
// 		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
// 	}), mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)

// 	err := awsCli.StartInstance(ctx, instanceId)
// 	require.NoError(t, err)

// 	mockClient.AssertExpectations(t)
// }

// func TestStopInstance(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	instanceId := "i-1234567890abcdef0"
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	mockClient.On("StopInstances", ctx, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
// 		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
// 	}), mock.Anything).Return(&ec2.StopInstancesOutput{}, nil)

// 	err := awsCli.StopInstance(ctx, instanceId)
// 	require.NoError(t, err)

// 	mockClient.AssertExpectations(t)
// }

// func TestFindInstances(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	instanceName := "instance-name"
// 	controllerID := "controllerID"
// 	tags := []types.Tag{
// 		{
// 			Key:   aws.String("tag:GARM_CONTROLLER_ID"),
// 			Value: &controllerID,
// 		},
// 		{
// 			Key:   aws.String("tag:Name"),
// 			Value: &instanceName,
// 		},
// 	}
// 	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
// 		return len(input.Filters) == 3
// 	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: aws.String("i-1234567890abcdef0"),
// 						Tags:       tags,
// 					},
// 					{
// 						InstanceId: aws.String("i-1234567890abcdef1"),
// 						Tags:       tags,
// 					},
// 				},
// 			},
// 		},
// 	}, nil)

// 	instances, err := awsCli.FindInstances(ctx, controllerID, instanceName)
// 	require.NoError(t, err)
// 	require.Len(t, instances, 2)
// 	require.Equal(t, tags, instances[0].Tags)
// 	require.Equal(t, tags, instances[1].Tags)

// 	mockClient.AssertExpectations(t)
// }

// func TestFindOneInstanceWithName(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	instanceName := "instance-name"
// 	controllerID := "controllerID"
// 	tags := []types.Tag{
// 		{
// 			Key:   aws.String("tag:GARM_CONTROLLER_ID"),
// 			Value: &controllerID,
// 		},
// 		{
// 			Key:   aws.String("tag:Name"),
// 			Value: &instanceName,
// 		},
// 	}
// 	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
// 		return len(input.Filters) == 3
// 	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: aws.String("i-1234567890abcdef0"),
// 						Tags:       tags,
// 					},
// 				},
// 			},
// 		},
// 	}, nil)

// 	instance, err := awsCli.FindOneInstance(ctx, controllerID, instanceName)
// 	require.NoError(t, err)
// 	require.Equal(t, tags, instance.Tags)

// 	mockClient.AssertExpectations(t)
// }

// func TestFindOneInstanceWithID(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	instanceId := "i-1234567890abcdef0"
// 	controllerID := "controllerID"
// 	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
// 		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
// 	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: &instanceId,
// 					},
// 				},
// 			},
// 		},
// 	}, nil)

// 	instance, err := awsCli.FindOneInstance(ctx, controllerID, instanceId)
// 	require.NoError(t, err)
// 	require.Equal(t, instanceId, *instance.InstanceId)

// 	mockClient.AssertExpectations(t)
// }

// func TestGetInstance(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	instanceID := "i-1234567890abcdef0"
// 	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
// 		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceID
// 	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						InstanceId: aws.String(instanceID),
// 					},
// 				},
// 			},
// 		},
// 	}, nil)

// 	instance, err := awsCli.GetInstance(ctx, instanceID)
// 	require.NoError(t, err)
// 	require.Equal(t, instanceID, *instance.InstanceId)
// }

// func TestTerminateInstance(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	poolID := "poolID"
// 	tags := []types.Tag{
// 		{
// 			Key:   aws.String("tag:GARM_POOL_ID"),
// 			Value: &poolID,
// 		},
// 	}
// 	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
// 		return len(input.Filters) == 2 && input.Filters[0].Name == aws.String("tag:GARM_POOL_ID")
// 	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
// 		Reservations: []types.Reservation{
// 			{
// 				Instances: []types.Instance{
// 					{
// 						Tags: tags,
// 					},
// 				},
// 			},
// 		},
// 	}, nil)

// 	mockClient.On("TerminateInstances", ctx, mock.MatchedBy(func(input *ec2.TerminateInstancesInput) bool {
// 		return len(input.InstanceIds) == 1
// 	}), mock.Anything).Return(&ec2.TerminateInstancesOutput{}, nil)

// 	err := awsCli.TerminateInstance(ctx, poolID)
// 	require.NoError(t, err)
// }

// func TestCreateRunningInstance(t *testing.T) {
// 	ctx := context.Background()
// 	cfg := &config.Config{
// 		Region:   "us-west-2",
// 		SubnetID: "subnet-1234567890abcdef0",
// 		Credentials: config.Credentials{
// 			CredentialType: config.AWSCredentialTypeStatic,
// 			StaticCredentials: config.StaticCredentials{
// 				AccessKeyID:     "AccessKeyID",
// 				SecretAccessKey: "SecretAccessKey",
// 				SessionToken:    "SessionToken",
// 			},
// 		},
// 	}
// 	mockClient := new(MockComputeClient)
// 	awsCli := &AwsCli{
// 		cfg:    cfg,
// 		client: mockClient,
// 	}
// 	instanceID := "i-1234567890abcdef0"
// 	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
// 		return params.RunnerApplicationDownload{
// 			OS:           aws.String("linux"),
// 			Architecture: aws.String("amd64"),
// 			DownloadURL:  aws.String("MockURL"),
// 			Filename:     aws.String("garm-runner"),
// 		}, nil
// 	}
// 	spec := &spec.RunnerSpec{
// 		Region: "us-west-2",
// 		Tools: params.RunnerApplicationDownload{
// 			OS:           aws.String("linux"),
// 			Architecture: aws.String("amd64"),
// 			DownloadURL:  aws.String("MockURL"),
// 			Filename:     aws.String("garm-runner"),
// 		},
// 		BootstrapParams: params.BootstrapInstance{
// 			Name:   "instance-name",
// 			OSType: "linux",
// 			Image:  "ami-12345678",
// 			Flavor: "t2.micro",
// 			PoolID: "poolID",
// 		},
// 		SubnetID:     "subnet-1234567890abcdef0",
// 		SSHKeyName:   aws.String("SSHKeyName"),
// 		ControllerID: "controllerID",
// 	}
// 	mockClient.On("RunInstances", ctx, mock.Anything, mock.Anything).Return(&ec2.RunInstancesOutput{
// 		Instances: []types.Instance{
// 			{
// 				InstanceId: aws.String(instanceID),
// 				KeyName:    aws.String("SSHKeyName"),
// 			},
// 		},
// 	}, nil)

// 	instance, err := awsCli.CreateRunningInstance(ctx, spec)
// 	require.NoError(t, err)
// 	require.Equal(t, instanceID, instance)
// }
