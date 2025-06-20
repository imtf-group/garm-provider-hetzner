package spec

import (
	"encoding/json"
	"testing"

	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/stretchr/testify/require"
)

func TestExtraSpecsFromBootstrapData(t *testing.T) {
	tests := []struct {
		name           string
		input          params.BootstrapInstance
		expectedOutput *extraSpecs
		errString      string
	}{
		{
			name: "test complete extraSpecs",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"placement_group": 444444, "firewalls": [222222, 333333], "networks": [111111], "location": "nbg1", "ssh_keys": [123456], "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==", "pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}, "datacenter": "nbg1-dc1"}`),
			},
			expectedOutput: &extraSpecs{
				Location:        hcloud.String("nbg1"),
				SSHKeys:         []int64{123456},
				Datacenter:      hcloud.String("nbg1-dc1"),
				PlacementGroup:  hcloud.Ptr(int64(444444)),
				Networks:        []int64{111111},
				Firewalls:       []int64{222222, 333333},
				DisableUpdates:  hcloud.Bool(true),
				EnableBootDebug: hcloud.Bool(true),
				ExtraPackages:   []string{"package1", "package2"},
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
				},
			},
			errString: "",
		},
		{
			name: "test no extraSpecs",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{}`),
			},
			expectedOutput: &extraSpecs{},
			errString:      "",
		},
		{
			name: "test invalid location",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"location": true}`),
			},
			expectedOutput: nil,
			errString:      "location: Invalid type. Expected: string, given: boolean",
		},
		{
			name: "test invalid ssh_key",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"ssh_keys": "not an array"}`),
			},
			expectedOutput: nil,
			errString:      "ssh_keys: Invalid type. Expected: array, given: string",
		},
		{
			name: "test invalid networks",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"networks": "not an array"}`),
			},
			expectedOutput: nil,
			errString:      "networks: Invalid type. Expected: array, given: string",
		},
		{
			name: "test invalid firewalls",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"firewalls": "not an array"}`),
			},
			expectedOutput: nil,
			errString:      "firewalls: Invalid type. Expected: array, given: string",
		},
		{
			name: "test invalid datacenter",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"datacenter": true}`),
			},
			expectedOutput: nil,
			errString:      "datacenter: Invalid type. Expected: string, given: boolean",
		},
		{
			name: "test invalid placement_group",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"placement_group": true}`),
			},
			expectedOutput: nil,
			errString:      "placement_group: Invalid type. Expected: integer, given: boolean",
		},
		{
			name: "test invalid disable_updates",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": "true"}`),
			},
			expectedOutput: nil,
			errString:      "disable_updates: Invalid type. Expected: boolean, given: string",
		},
		{
			name: "test invalid enable_boot_debug",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"enable_boot_debug": "true"}`),
			},
			expectedOutput: nil,
			errString:      "enable_boot_debug: Invalid type. Expected: boolean, given: string",
		},
		{
			name: "test invalid extra_packages",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_packages": "package1"}`),
			},
			expectedOutput: nil,
			errString:      "extra_packages: Invalid type. Expected: array, given: string",
		},
		{
			name: "test invalid runner_install_template",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"runner_install_template": 123}`),
			},
			expectedOutput: nil,
			errString:      "runner_install_template: Invalid type. Expected: string, given: integer",
		},
		{
			name: "test invalid pre_install_scripts",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"pre_install_scripts": "setup.sh"}`),
			},
			expectedOutput: nil,
			errString:      "pre_install_scripts: Invalid type. Expected: object, given: string",
		},
		{
			name: "test invalid extra_context",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_context": 123}`),
			},
			expectedOutput: nil,
			errString:      "extra_context: Invalid type. Expected: object, given: integer",
		},
		{
			name: "test invalid property",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"invalid": "invalid"}`),
			},
			expectedOutput: nil,
			errString:      "Additional property invalid is not allowed",
		},
		{
			name: "test invalid json",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": }`),
			},
			expectedOutput: nil,
			errString:      "failed to validate extra specs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := newExtraSpecsFromBootstrapData(tt.input)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
			require.Equal(t, tt.expectedOutput, output)
		})
	}
}

func TestGetRunnerSpecFromBootstrapParams(t *testing.T) {
	Mocktools := params.RunnerApplicationDownload{
		OS:           hcloud.String("linux"),
		Architecture: hcloud.String("amd64"),
		DownloadURL:  hcloud.String("MockURL"),
		Filename:     hcloud.String("garm-runner"),
	}
	DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return Mocktools, nil
	}

	data := params.BootstrapInstance{
		Name:       "mock-name",
		Image:      "ubuntu-24.04",
		ExtraSpecs: json.RawMessage(`{"disable_ipv6": true, "disable_ipv4": false, "placement_group": 444444, "firewalls": [222222, 333333], "networks": [111111], "location": "nbg1", "ssh_keys": [123456], "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==", "pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}, "datacenter": "nbg1-dc1"}`),
	}

	config := &config.Config{
		Location: "nbg1",
		Token:    "hcloud-token",
	}
	expectedRunnerSpec := &RunnerSpec{
		Location:        "nbg1",
		ExtraPackages:   []string{"package1", "package2"},
		SSHKeys:         []int64{123456},
		Datacenter:      "nbg1-dc1",
		PlacementGroup:  444444,
		Networks:        []int64{111111},
		Firewalls:       []int64{222222, 333333},
		DisableUpdates:  true,
		EnableBootDebug: true,
		Tools:           Mocktools,
		ControllerID:    "controller_id",
		BootstrapParams: data,
		DisableIPv4:     false,
		DisableIPv6:     true,
	}

	runnerSpec, err := GetRunnerSpecFromBootstrapParams(config, data, "controller_id")
	require.NoError(t, err)
	require.Equal(t, expectedRunnerSpec, runnerSpec)
}

func TestRunnerSpecValidate(t *testing.T) {
	tests := []struct {
		name      string
		spec      *RunnerSpec
		errString string
	}{
		{
			name:      "empty runner spec",
			spec:      &RunnerSpec{},
			errString: "missing region",
		},
		{
			name: "missing name",
			spec: &RunnerSpec{
				Location: "location",
				BootstrapParams: params.BootstrapInstance{
					Image: "ubuntu-24.04",
				},
			},
			errString: "missing bootstrap params",
		},
		{
			name: "missing image",
			spec: &RunnerSpec{
				Location: "location",
				BootstrapParams: params.BootstrapInstance{
					Name: "name",
				},
			},
			errString: "missing bootstrap params",
		},
		{
			name: "valid runner spec",
			spec: &RunnerSpec{
				Location:        "nbg1",
				ExtraPackages:   []string{"package1", "package2"},
				SSHKeys:         []int64{123456},
				Datacenter:      "nbg1-dc1",
				PlacementGroup:  444444,
				Networks:        []int64{111111},
				Firewalls:       []int64{222222, 333333},
				DisableUpdates:  true,
				EnableBootDebug: true,
				Tools: params.RunnerApplicationDownload{
					OS:           hcloud.String("linux"),
					Architecture: hcloud.String("amd64"),
					DownloadURL:  hcloud.String("MockURL"),
					Filename:     hcloud.String("garm-runner"),
				},
				ControllerID: "controller_id",
				BootstrapParams: params.BootstrapInstance{
					Name:  "name",
					Image: "ubuntu-24.04",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestMergeExtraSpecs(t *testing.T) {
	tests := []struct {
		name     string
		spec     *RunnerSpec
		extra    *extraSpecs
		expected *RunnerSpec
	}{
		{
			name: "empty extra specs",
			spec: &RunnerSpec{
				Location: "location",
			},
			extra:    &extraSpecs{},
			expected: &RunnerSpec{Location: "location"},
		},
		{
			name: "valid extra specs",
			spec: &RunnerSpec{
				Location: "location",
			},
			extra: &extraSpecs{
				Location:        hcloud.String("nbg1"),
				SSHKeys:         []int64{123456},
				Datacenter:      hcloud.String("nbg1-dc1"),
				PlacementGroup:  hcloud.Ptr(int64(444444)),
				Networks:        []int64{111111},
				Firewalls:       []int64{222222, 333333},
				DisableUpdates:  hcloud.Bool(true),
				EnableBootDebug: hcloud.Bool(true),
			},
			expected: &RunnerSpec{
				Location:        "nbg1",
				SSHKeys:         []int64{123456},
				Datacenter:      "nbg1-dc1",
				PlacementGroup:  444444,
				Networks:        []int64{111111},
				Firewalls:       []int64{222222, 333333},
				DisableUpdates:  true,
				EnableBootDebug: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.spec.MergeExtraSpecs(tt.extra)
			require.Equal(t, tt.expected, tt.spec)
		})
	}
}
