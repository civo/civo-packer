// The Civo package contains a packer.Builder implementation
// that builds Civo images (snapshots).

package civo

import (
	"context"
	"fmt"
	"log"

	"github.com/civo/civogo"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// BuilderID unique id for the builder
const BuilderID = "pearkes.civo"

// Builder is a struc
type Builder struct {
	config Config
	runner multistep.Runner
}

// ConfigSpec ...
func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

// Prepare ...
func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, nil, nil
}

// Run ...
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	client, err := civogo.NewClient(b.config.APIToken)
	if err != nil {
		return nil, fmt.Errorf("civo: %s", err)
	}

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("civo_%s.pem", b.config.PackerBuildName),
		},
		new(stepCreateInstance),
		new(stepInstanceInfo),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.Host(), "instance_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(common.StepProvision),
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		new(stepShutdown),
		new(stepPowerOff),
		&stepSnapshot{
			snapshotTimeout: b.config.SnapshotTimeout,
		},
	}

	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("snapshot_name"); !ok {
		log.Println("Failed to find snapshot_name in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		SnapshotName: state.Get("snapshot_name").(string),
		SnapshotID:   state.Get("snapshot_id").(string),
		RegionNames:  state.Get("regions").([]string),
		Client:       client,
	}

	return artifact, nil
}
