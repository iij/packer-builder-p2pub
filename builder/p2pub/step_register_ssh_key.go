package p2pub

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

type stepRegisterSSHKey struct {
}

func (s *stepRegisterSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	storageServiceCode := state.Get("StorageServiceCode").(string)
	OSType := state.Get("OSType").(string)

	if config.RootSSHKey != "" && OSType == "Linux" {

		ui.Say(fmt.Sprintf("Registering SSH public key ..."))

		keyArgs := protocol.PublicKeyAdd{
			GisServiceCode:     config.GisServiceCode,
			StorageServiceCode: storageServiceCode,
			PublicKey:          config.RootSSHKey,
		}
		keyResp := protocol.PublicKeyAddResponse{}
		if err := p2pubapi.Call(*api, keyArgs, &keyResp); err != nil {
			ui.Error(fmt.Sprintf("Failed to set ssh public key. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}

		if err := p2pubapi.WaitSystemStorage(api, config.GisServiceCode, storageServiceCode,
			p2pubapi.InService, p2pubapi.NotAttached, 10*time.Minute); err != nil {
			ui.Error(fmt.Sprintf("System storage could not be ready. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepRegisterSSHKey) Cleanup(state multistep.StateBag) {
}
