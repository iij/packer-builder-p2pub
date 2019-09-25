package p2pub

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

type stepCreateStorageArchive struct {
}

func (s *stepCreateStorageArchive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Preparing storage archive ..."))

	checkArgs := protocol.P2PUBContractGetForSA{
		GisServiceCode: config.GisServiceCode,
	}
	checkResp := protocol.P2PUBContractGetForSAResponse{}
	if err := p2pubapi.Call(*api, checkArgs, &checkResp); err != nil {
		ui.Error(fmt.Sprintf("Failed to check storage archive's status. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if checkResp.StorageArchive.ServiceCode != "" {
		state.Put("IarServiceCode", checkResp.StorageArchive.ServiceCode)
	} else {
		createArgs := protocol.StorageArchiveAdd{
			GisServiceCode: config.GisServiceCode,
			ArchiveSize:    "100",
		}
		createResp := protocol.StorageArchiveAddResponse{}
		if err := p2pubapi.Call(*api, createArgs, &createResp); err != nil {
			ui.Error(fmt.Sprintf("Failed to prepare storage archive. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("IarServiceCode", createResp.ServiceCode)
	}

	ui.Say(fmt.Sprintf("Using storage archive '%s'", state.Get("IarServiceCode").(string)))

	return multistep.ActionContinue
}

func (s *stepCreateStorageArchive) Cleanup(state multistep.StateBag) {
}
