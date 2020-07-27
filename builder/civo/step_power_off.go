package civo

import (
	"context"
	"fmt"
	"log"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instanceID := state.Get("instance_id").(string)

	instance, err := client.GetInstance(instanceID)
	if err != nil {
		err := fmt.Errorf("Error checking instance state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instance.Status == "SHUTOFF" {
		// instance is already off, don't do anything
		return multistep.ActionContinue
	}

	// Pull the plug on the instance
	ui.Say("Forcefully shutting down instance...")
	_, err = client.StopInstance(instanceID)
	if err != nil {
		err := fmt.Errorf("Error powering off instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for poweroff event to complete...")
	err = waitForInstanceState("SHUTOFF", instanceID, client, c.StateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state multistep.StateBag) {
	// no cleanup
}
