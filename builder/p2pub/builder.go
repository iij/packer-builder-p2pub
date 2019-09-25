package p2pub

import (
	"context"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
)

// Builder ...
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare ...
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	conf, warn, err := newConfig(raws...)
	if err != nil {
		return warn, err
	}
	b.config = conf
	return warn, nil
}

// Run ...
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	api := p2pubapi.NewAPI(b.config.APIAccessKey, b.config.APISecretKey)

	state := &multistep.BasicStateBag{}
	state.Put("config", b.config)
	state.Put("api", api)
	state.Put("ui", ui)
	state.Put("hook", hook)

	steps := []multistep.Step{
		&stepCreateSSHKey{},
		&stepCreateStorageArchive{},
		&stepCreateSystemStorage{},
		&stepRestoreBaseImage{},
		&stepRegisterSSHKey{},
		&stepCreateVirtualServer{},
		&stepAttachSystemStorage{},
		&stepAllocateIPAddress{},
		&stepBoot{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
			Host:      communicator.CommHost(b.config.Comm.SSHHost, "IpAddress"),
		},
		&common.StepProvision{},
		&stepHalt{},
		&stepArchive{},
	}

	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		GisServiceCode: b.config.GisServiceCode,
		IarServiceCode: state.Get("IarServiceCode").(string),
		ImageID:        state.Get("ImageId").(string),
		api:            api,
	}

	return artifact, nil
}
