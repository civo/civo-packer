package civo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepShutdown struct{}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instanceID := state.Get("instance_id").(string)

	// Gracefully power off the instance. We have to retry this a number
	// of times because sometimes it says it completed when it actually
	// did absolutely nothing (*ALAKAZAM!* magic!). We give up after
	// a pretty arbitrary amount of time.
	ui.Say("Gracefully shutting down instance...")
	_, err := client.StopInstance(instanceID)
	if err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error shutting down instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// A channel we use as a flag to end our goroutines
	done := make(chan struct{})
	shutdownRetryDone := make(chan struct{})

	// Make sure we wait for the shutdown retry goroutine to end
	// before moving on.
	defer func() {
		close(done)
		<-shutdownRetryDone
	}()

	// Start a goroutine that just keeps trying to shut down the
	// instance.
	go func() {
		defer close(shutdownRetryDone)

		for attempts := 2; attempts > 0; attempts++ {
			log.Printf("Shutdowninstance attempt #%d...", attempts)
			_, err := client.StopInstance(instanceID)
			if err != nil {
				log.Printf("Shutdown retry error: %s", err)
			}

			select {
			case <-done:
				return
			case <-time.After(20 * time.Second):
				// Retry!
			}
		}
	}()

	err = waitForInstanceState("SHUTOFF", instanceID, client, c.StateTimeout)
	if err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error shutting down instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}
