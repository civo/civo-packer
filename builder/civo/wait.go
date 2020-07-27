package civo

import (
	"fmt"
	"log"
	"time"

	"github.com/civo/civogo"
)

// waitForinstanceState simply blocks until the instance is in
// a state we expect, while eventually timing out.
func waitForInstanceState(
	desiredState string, instanceID string, client *civogo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts++

			log.Printf("Checking instance status... (attempt: %d)", attempts)
			instance, err := client.GetInstance(instanceID)
			if err != nil {
				result <- err
				return
			}

			if instance.Status == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for instance to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for instance to become '%s'", desiredState)
		return err
	}
}

// waitForActionState simply blocks until the instance action is in
// a state we expect, while eventually timing out.
func waitForActionState(desiredState string, action string, client *civogo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts++

			log.Printf("Checking action status... (attempt: %d)", attempts)
			action, err := client.FindSnapshot(action)
			if err != nil {
				result <- err
				return
			}

			if action.State == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for action to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for action to become '%s'", desiredState)
		return err
	}
}

// WaitForImageState simply blocks until the image action is in
// a state we expect, while eventually timing out.
func WaitForImageState(
	desiredState string, imageID string, client *civogo.Client, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts++

			log.Printf("Checking action status... (attempt: %d)", attempts)
			action, err := client.FindSnapshot(imageID)
			if err != nil {
				result <- err
				return
			}

			if action.State == desiredState {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for image transfer to become %s", timeout/time.Second, desiredState)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for image transfer to become '%s'", desiredState)
		return err
	}
}
