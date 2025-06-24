package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudbase/garm-provider-common/execution"
	"github.com/imtf-group/garm-provider-hetzner/provider"
)

var signals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting environment: %q", err)
		os.Exit(1)
	}
	prov, err := provider.NewHcloudProvider(ctx, executionEnv.ProviderConfigFile, executionEnv.ControllerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %q", err)
		os.Exit(1)
	}
	result, err := executionEnv.Run(ctx, prov)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %+v\n", err)
		os.Exit(1)
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result) //nolint:errcheck
	}
}
