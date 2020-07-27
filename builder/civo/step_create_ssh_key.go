package civo

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/civo/civogo"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string

	keyID string
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Creating temporary ssh key for instance...")

	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		err := fmt.Errorf("error generating RSA key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// ASN.1 DER encoded form
	privder := x509.MarshalPKCS1PrivateKey(priv)
	privblk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privder,
	}

	// Set the private key in the config for later
	c.Comm.SSHPrivateKey = pem.EncodeToMemory(&privblk)

	// Marshal the public key into SSH compatible format
	// TODO properly handle the public key error
	pub, _ := ssh.NewPublicKey(&priv.PublicKey)
	pubsshformat := string(ssh.MarshalAuthorizedKey(pub))

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	key, err := client.NewSSHKey(name, pubsshformat)
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyID = key.ID

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future and Save the keys in the state bag
	state.Put("ssh_key_id", key.ID)
	// state.Put("ssh_private_key", string(pem.EncodeToMemory(&privblk)))
	// state.Put("ssh_public_key", string(ssh.MarshalAuthorizedKey(pub)))

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&privblk)); err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("Error setting permissions of debug key: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyID == "" {
		return
	}

	client := state.Get("client").(*civogo.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.DeleteSSHKey(s.keyID)
	if err != nil {
		log.Printf("Error cleaning up ssh key: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", err))
	}
}
