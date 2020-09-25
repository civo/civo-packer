package civo

import (
	"context"
	"fmt"
	"log"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateInstance struct {
	instanceID string
}

// Run function to run create a instance
func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	sshKeyID := state.Get("ssh_key_id").(string)

	// Create the instance based on configuration
	ui.Say("Creating instance...")

	template, err := client.FindTemplate(c.Template)
	if err != nil {
		ui.Error(err.Error())
	}

	network, _ := client.GetDefaultNetwork()

	InstanceConfig := &civogo.InstanceConfig{
		Hostname:         c.InstanceName,
		PublicIPRequired: c.PublicNetworking,
		Region:           c.Region,
		NetworkID:        network.ID,
		InitialUser:      c.Comm.SSHUsername,
		Size:             c.Size,
		TemplateID:       template.ID,
		SSHKeyID:         sshKeyID,
	}

	log.Printf("[DEBUG] Instance create paramaters: %+v", InstanceConfig)

	instance, err := client.CreateInstance(InstanceConfig)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.instanceID = instance.ID

	// Store the instance id for later
	state.Put("instance_id", instance.ID)

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	// If the instanceid isn't there, we probably never created it
	if s.instanceID == "" {
		return
	}

	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)

	// Destroy the instance we just created
	ui.Say("Destroying instance...")
	_, err := client.DeleteInstance(s.instanceID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying instance. Please destroy it manually: %s", err))
	}
}
