package spec

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-common/util"
	"github.com/imtf-group/garm-provider-hetzner/config"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

type ToolFetchFunc func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error)

var DefaultToolFetch ToolFetchFunc = util.GetTools

func generateJSONSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
	}
	schema := reflector.Reflect(extraSpecs{})

	return schema
}

func jsonSchemaValidation(schema json.RawMessage) error {
	jsonSchema := generateJSONSchema()
	schemaLoader := gojsonschema.NewGoLoader(jsonSchema)
	extraSpecsLoader := gojsonschema.NewBytesLoader(schema)
	result, err := gojsonschema.Validate(schemaLoader, extraSpecsLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("schema validation failed: %s", result.Errors())
	}
	return nil
}

func newExtraSpecsFromBootstrapData(data params.BootstrapInstance) (*extraSpecs, error) {
	spec := &extraSpecs{}

	if err := jsonSchemaValidation(data.ExtraSpecs); err != nil {
		return nil, fmt.Errorf("failed to validate extra specs: %w", err)
	}

	if len(data.ExtraSpecs) > 0 {
		if err := json.Unmarshal(data.ExtraSpecs, spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra specs: %w", err)
		}
	}

	return spec, nil
}

type extraSpecs struct {
	Location        *string  `json:"location,omitempty" jsonschema:"description=Location where to create the server."`
	SSHKeys         []int64  `json:"ssh_keys,omitempty" jsonschema:"description=ID of SSH keys to use for the instance."`
	Datacenter      *string  `json:"datacenter,omitempty" jsonschema:"description=Datacenter where to create the server."`
	PlacementGroup  *int64   `json:"placement_group,omitempty" jsonschema:"description=ID of the placement Group where the Server should be in."`
	Networks        []int64  `json:"networks,omitempty" jsonschema:"description=Network IDs which should be attached to the Server private network interface."`
	Firewalls       []int64  `json:"firewalls,omitempty" jsonschema:"description=Firewall IDs which should be applied on the Server's public network interface."`
	DisableUpdates  *bool    `json:"disable_updates,omitempty" jsonschema:"description=Disable automatic updates on the VM."`
	EnableBootDebug *bool    `json:"enable_boot_debug,omitempty" jsonschema:"description=Enable boot debug on the VM."`
	DisableIPv4     *bool    `json:"disable_ipv4,omitempty" jsonschema:"description=Disable public IPv4."`
	DisableIPv6     *bool    `json:"disable_ipv6,omitempty" jsonschema:"description=Disable public IPv6."`
	ExtraPackages   []string `json:"extra_packages,omitempty" jsonschema:"description=Extra packages to install on the VM."`
	cloudconfig.CloudConfigSpec
}

func GetRunnerSpecFromBootstrapParams(cfg *config.Config, data params.BootstrapInstance, controllerID string) (*RunnerSpec, error) {
	tools, err := DefaultToolFetch(data.OSType, data.OSArch, data.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %s", err)
	}

	extraSpecs, err := newExtraSpecsFromBootstrapData(data)
	if err != nil {
		return nil, fmt.Errorf("error loading extra specs: %w", err)
	}

	spec := &RunnerSpec{
		Location:        cfg.Location,
		ExtraPackages:   extraSpecs.ExtraPackages,
		Tools:           tools,
		BootstrapParams: data,
		ControllerID:    controllerID,
	}

	spec.MergeExtraSpecs(extraSpecs)

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("error validating spec: %w", err)
	}

	return spec, nil
}

type RunnerSpec struct {
	Location        string
	DisableUpdates  bool
	ExtraPackages   []string
	Datacenter      string
	EnableBootDebug bool
	Tools           params.RunnerApplicationDownload
	BootstrapParams params.BootstrapInstance
	SSHKeys         []int64
	PlacementGroup  int64
	Networks        []int64
	Firewalls       []int64
	DisableIPv4     bool
	DisableIPv6     bool
	ControllerID    string
}

func (r *RunnerSpec) Validate() error {
	if r.Location == "" {
		return fmt.Errorf("missing region")
	}
	if r.BootstrapParams.Name == "" {
		return fmt.Errorf("missing bootstrap params")
	}
	if r.BootstrapParams.Image == "" {
		return fmt.Errorf("missing bootstrap params")
	}
	return nil
}

func (r *RunnerSpec) MergeExtraSpecs(extraSpecs *extraSpecs) {
	if extraSpecs.SSHKeys != nil {
		r.SSHKeys = extraSpecs.SSHKeys
	}

	if extraSpecs.Location != nil {
		r.Location = *extraSpecs.Location
	}

	if extraSpecs.Datacenter != nil {
		r.Datacenter = *extraSpecs.Datacenter
	}

	if extraSpecs.PlacementGroup != nil {
		r.PlacementGroup = *extraSpecs.PlacementGroup
	}

	if extraSpecs.Networks != nil {
		r.Networks = extraSpecs.Networks
	}

	if extraSpecs.Firewalls != nil {
		r.Firewalls = extraSpecs.Firewalls
	}

	if extraSpecs.DisableUpdates != nil {
		r.DisableUpdates = *extraSpecs.DisableUpdates
	}

	if extraSpecs.EnableBootDebug != nil {
		r.EnableBootDebug = *extraSpecs.EnableBootDebug
	}

	if extraSpecs.DisableIPv4 != nil {
		r.DisableIPv4 = *extraSpecs.DisableIPv4
	}

	if extraSpecs.DisableIPv6 != nil {
		r.DisableIPv6 = *extraSpecs.DisableIPv6
	}
}

func (r *RunnerSpec) ComposeUserData() (string, error) {
	bootstrapParams := r.BootstrapParams
	bootstrapParams.UserDataOptions.DisableUpdatesOnBoot = r.DisableUpdates
	bootstrapParams.UserDataOptions.ExtraPackages = r.ExtraPackages
	bootstrapParams.UserDataOptions.EnableBootDebug = r.EnableBootDebug
	switch bootstrapParams.OSType {
	case params.Linux:
		udata, err := cloudconfig.GetCloudConfig(bootstrapParams, r.Tools, bootstrapParams.Name)
		if err != nil {
			return "", fmt.Errorf("failed to generate userdata: %w", err)
		}
		asBase64 := base64.StdEncoding.EncodeToString([]byte(udata))
		return asBase64, nil
	case params.Windows:
		udata, err := cloudconfig.GetCloudConfig(bootstrapParams, r.Tools, bootstrapParams.Name)
		if err != nil {
			return "", fmt.Errorf("failed to generate userdata: %w", err)
		}
		wrapped := fmt.Sprintf("<powershell>%s</powershell>", udata)
		asBase64 := base64.StdEncoding.EncodeToString([]byte(wrapped))
		return asBase64, nil
	}
	return "", fmt.Errorf("unsupported OS type for cloud config: %s", bootstrapParams.OSType)
}
