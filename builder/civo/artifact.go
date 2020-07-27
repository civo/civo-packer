package civo

import (
	"fmt"
	"log"
	"strings"

	"github.com/civo/civogo"
)

// Artifact ...
type Artifact struct {
	// The name of the snapshot
	SnapshotName string
	// The ID of the image
	SnapshotID string
	// The name of the region
	RegionNames []string
	// The client for making API calls
	Client *civogo.Client
}

// BuilderId ...
func (*Artifact) BuilderId() string {
	return BuilderID
}

// Files ...
func (*Artifact) Files() []string {
	// No files with Civo
	return nil
}

// Id ...
func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", strings.Join(a.RegionNames[:], ","), a.SnapshotID)
}

// String ...
func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%s' (ID: %s) in regions '%s'", a.SnapshotName, a.SnapshotID, strings.Join(a.RegionNames[:], ","))
}

// State ...
func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy ...
func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.SnapshotID, a.SnapshotName)
	_, err := a.Client.DeleteSnapshot(a.SnapshotID)
	return err
}
