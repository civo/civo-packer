package civo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepSnapshot struct {
	snapshotTimeout time.Duration
}

func (s *stepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	instanceID := state.Get("instance_id").(string)
	var snapshotRegions []string
	var snapShotConfig = &civogo.SnapshotConfig{
		InstanceID: instanceID,
		Safe:       true,
		Cron:       "",
	}

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	action, err := client.CreateSnapshot(c.SnapshotName, snapShotConfig)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// With the pending state over, verify that we're in the active state
	// because action can take a long time and may depend on the size of the final snapshot,
	// the timeout is parameterized
	ui.Say("Waiting for snapshot to complete...")
	if err := waitForActionState("complete", action.ID, client, s.snapshotTimeout); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the instance to become unlocked first. For snapshots
	// this can end up taking quite a long time, so we hardcode this to
	// 20 minutes.
	// if err := waitForInstanceUnlocked(client, instanceID, 20*time.Minute); err != nil {
	// 	// If we get an error the first time, actually report it
	// 	err := fmt.Errorf("Error shutting down instance: %s", err)
	// 	state.Put("error", err)
	// 	ui.Error(err.Error())
	// 	return multistep.ActionHalt
	// }

	log.Printf("Looking up snapshot ID for snapshot: %s", c.SnapshotName)
	images, err := client.FindSnapshot(c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error looking up snapshot ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var imageID string
	if images.ID != "" {
		imageID = images.ID
	} else {
		err := errors.New("Couldn't find snapshot to get the image ID. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	snapshotRegions = append(snapshotRegions, c.Region)

	log.Printf("Snapshot image ID: %s", imageID)
	state.Put("snapshot_id", imageID)
	state.Put("snapshot_name", c.SnapshotName)
	state.Put("regions", snapshotRegions)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
