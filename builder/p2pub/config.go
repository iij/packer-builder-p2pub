package p2pub

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config ...
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	// API Key for IIJ GIO P2 (required)
	// (https://manual.iij.jp/p2/pubapi/59950199.html)
	APIAccessKey string `mapstructure:"access_key_id"`
	APISecretKey string `mapstructure:"secret_access_key"`

	// Servicecode of the P2 contract used to built and images save in (required)
	GisServiceCode string `mapstructure:"gis_service_code"`

	// System storage type of built image (required)
	// List of storage types: https://manual.iij.jp/p2/pubapi/59949023.html
	StorageType string `mapstructure:"storage_type"`

	// Virtual Server type use to build image (optional. default: "VB0-1")
	VMType string `mapstructure:"server_type"`

	// Another image built images are based on (optional)
	BaseImage struct {
		GisServiceCode string `mapstructure:"gis_service_code"`
		IarServiceCode string `mapstructure:"iar_service_code"`
		ImageID        string `mapstructure:"image_id"`
	} `mapstructure:"base_image"`

	// SSH connection settings
	// Public key ** for "root" ** (optional)
	RootSSHKey string `mapstructure:"root_ssh_key"`

	// The label set to built image (optional)
	Label string `mapstructure:"label"`

	// If this is set true, P2 PUB builder connects VMs through Standard Private Network. (optional)
	// By default, this builder uses the Internet.
	DisableGlobalAddress bool `mapstructure:"disable_global_address"`
}

func newConfig(raws ...interface{}) (*Config, []string, error) {

	conf := new(Config)
	warns := []string{}

	decodeOpts := &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &conf.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"label",
			},
		},
	}

	if err := config.Decode(conf, decodeOpts, raws...); err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	if strings.Contains(conf.StorageType, "WINDOWS") {
		if conf.BaseImage.GisServiceCode == "" ||
			conf.BaseImage.IarServiceCode == "" ||
			conf.BaseImage.ImageID == "" {
			packer.MultiErrorAppend(errs, fmt.Errorf("Base image is not specified. \n You can connect only by using the console in P2 control panel when using vanilla Windows images."))
		}
	} else {
		if conf.Comm.SSHUsername == "" {
			// treated as root
			conf.Comm.SSHUsername = "root"
			if conf.RootSSHKey != "" && conf.Comm.SSHPrivateKeyFile == "" {
				packer.MultiErrorAppend(errs, fmt.Errorf("SSH private key must be specified"))
			}
		} else if conf.Comm.SSHUsername != "root" {
			if conf.Comm.SSHPrivateKeyFile == "" {
				packer.MultiErrorAppend(errs, fmt.Errorf("SSH private key must be specified"))
			}
			if conf.RootSSHKey != "" {
				warns = append(warns, "root_ssh_key is not affected while build")
			}
		}
	}

	if conf.VMType == "" {
		conf.VMType = "VB0-1"
	}

	if es := conf.Comm.Prepare(&conf.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return conf, nil, nil
}
