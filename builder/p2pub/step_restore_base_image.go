package p2pub

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

type stepRestoreBaseImage struct {
}

func (s *stepRestoreBaseImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	storageServiceCode := state.Get("StorageServiceCode").(string)

	// Base image is not specified. skip
	if config.BaseImage.GisServiceCode == "" &&
		config.BaseImage.IarServiceCode == "" &&
		config.BaseImage.ImageID == "" {
		return multistep.ActionContinue
	}

	// In case of Type-X System Storage, restoring is already completed in create step. skip
	if strings.HasPrefix(storageServiceCode, "ica") {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Restoring base image ..."))

	args := protocol.Restore{
		GisServiceCode:     config.GisServiceCode,
		StorageServiceCode: storageServiceCode,
		IarServiceCode:     config.BaseImage.IarServiceCode,
		ImageId:            config.BaseImage.ImageID,
		Image:              "Archive",
	}
	resp := protocol.RestoreResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(fmt.Sprintf("Failed to restore base image. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if err := p2pubapi.WaitSystemStorage(api, config.GisServiceCode, storageServiceCode,
		p2pubapi.InService, p2pubapi.NotAttached, 10*time.Minute); err != nil {
		ui.Error(fmt.Sprintf("System storage could not be ready. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRestoreBaseImage) Cleanup(state multistep.StateBag) {
}
