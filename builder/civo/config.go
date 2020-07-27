//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package civo

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

// Config to teh packer
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	// The client TOKEN to use to access your account. It
	// can also be specified via environment variable CIVO_TOKEN, if
	// set.
	APIToken string `mapstructure:"api_token" required:"true"`
	// The name (or slug) of the region to launch the instance
	// in. Consequently, this is the region where the snapshot will be available.
	Region string `mapstructure:"region" required:"true"`
	// The name (or slug) of the instance size to use. See
	Size string `mapstructure:"size" required:"true"`
	// The name (or slug) of the base image to use. This is the
	// image that will be used to launch a new instance and provision it.
	Template string `mapstructure:"template" required:"true"`
	// Set to true to enable private networking
	// for the instance being created. This defaults to true.
	PublicNetworking string `mapstructure:"private_networking" required:"false"`
	// The name of the resulting snapshot that will
	// appear in your account. Defaults to `packer-{{timestamp}}` (see
	// configuration templates for more info).
	SnapshotName string `mapstructure:"snapshot_name" required:"false"`
	// The regions of the resulting
	// snapshot that will appear in your account.
	SnapshotRegions []string `mapstructure:"snapshot_regions" required:"false"`
	// The time to wait, as a duration string, for a
	// instance to enter a desired state (such as "active") before timing out. The
	// default state timeout is "6m".
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`
	// How long to wait for an image to be published to the shared image
	// gallery before timing out. If your Packer build is failing on the
	// Publishing to Shared Image Gallery step with the error `Original Error:
	// context deadline exceeded`, but the image is present when you check your
	// Azure dashboard, then you probably need to increase this timeout from
	// its default of "60m" (valid time units include `s` for seconds, `m` for
	// minutes, and `h` for hours.)
	SnapshotTimeout time.Duration `mapstructure:"snapshot_timeout" required:"false"`
	// The name assigned to the instance. Civo sets the hostname of the machine to this value.
	InstanceName string `mapstructure:"instance_name" required:"false"`

	ctx interpolate.Context
}

// Prepare function to prepare the builder
func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if c.APIToken == "" {
		// Default to environment variable for api_token, if it exists
		c.APIToken = os.Getenv("CIVO_TOKEN")
	}
	if c.SnapshotName == "" {
		def, err := interpolate.Render("civo-packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to civo-packer-{{ unix timestamp (utc) }}
		c.SnapshotName = def
	}

	if c.InstanceName == "" {
		// Default to packer-[time-ordered-uuid]
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.StateTimeout == 0 {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for instance to become active
		c.StateTimeout = 6 * time.Minute
	}

	if c.SnapshotTimeout == 0 {
		// Default to 60 minutes timeout, waiting for snapshot action to finish
		c.SnapshotTimeout = 60 * time.Minute
	}

	if c.PublicNetworking == "" {
		c.PublicNetworking = "true"
	}

	var errs *packer.MultiError

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.APIToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("api_token for auth must be specified"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.Size == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("size is required"))
	}

	if c.Template == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("template is required"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(c.APIToken)
	return nil, nil
}
