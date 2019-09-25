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

type stepCreateSystemStorage struct {
}

func (s *stepCreateSystemStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Creating system storage ..."))

	encryption := "No"
	if strings.HasPrefix(config.StorageType, "SX") {
		encryption = "Yes"
	}

	args := protocol.SystemStorageAdd{
		GisServiceCode: config.GisServiceCode,
		Type:           config.StorageType,
		Encryption:     encryption,
	}

	if config.BaseImage.ImageID != "" && strings.HasPrefix(config.BaseImage.ImageID, "ica") {
		args.ImageId = config.BaseImage.ImageID
		args.SourceServiceCode = config.BaseImage.IarServiceCode
	}

	resp := protocol.SystemStorageAddResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(fmt.Sprintf("Failed to create system storage. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("StorageServiceCode", resp.ServiceCode)
	state.Put("OSType", resp.OSType)

	if err := p2pubapi.WaitSystemStorage(api, config.GisServiceCode, resp.ServiceCode,
		p2pubapi.InService, p2pubapi.NotAttached, 10*time.Minute); err != nil {
		ui.Error(fmt.Sprintf("System storage could not be ready. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Using system storage '%s'", state.Get("StorageServiceCode").(string)))

	return multistep.ActionContinue
}

func (s *stepCreateSystemStorage) Cleanup(state multistep.StateBag) {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say("Removing system storage ...")

	serviceCode := state.Get("StorageServiceCode").(string)

	if err := p2pubapi.WaitSystemStorage(api, config.GisServiceCode, serviceCode,
		p2pubapi.InService, p2pubapi.NotAttached, 10*time.Minute); err != nil {
		ui.Error(err.Error())
		ui.Say(fmt.Sprintf("Failed to remove system storage '%s'. remove it manually.", serviceCode))
		return
	}

	args := protocol.SystemStorageCancel{
		GisServiceCode:     config.GisServiceCode,
		StorageServiceCode: serviceCode,
	}
	resp := protocol.SystemStorageCancelResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(err.Error())
		ui.Say(fmt.Sprintf("Failed to remove system storage '%s'. remove it manually.", serviceCode))
	}
}
