package civo

import (
	"context"
	"fmt"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	instanceID := state.Get("instance_id").(string)

	ui.Say("Waiting for instance to become active...")

	err := waitForInstanceState("ACTIVE", instanceID, client, c.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	instance, err := client.GetInstance(instanceID)
	if err != nil {
		err := fmt.Errorf("Error retrieving instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Verify we have an IPv4 address
	invalid := instance.PublicIP == ""
	if invalid {
		err := fmt.Errorf("IPv4 address not found for instance")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Find a public IPv4 network
	foundNetwork := false
	if instance.PublicIP != "" {
		state.Put("instance_ip", instance.PublicIP)
		foundNetwork = true
	}

	if !foundNetwork {
		err := fmt.Errorf("Count not find a public IPv4 address for this instance")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
