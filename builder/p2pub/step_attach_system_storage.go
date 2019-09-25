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

type stepAttachSystemStorage struct {
}

func (s *stepAttachSystemStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ivm := state.Get("IvmServiceCode").(string)
	storageServiceCode := state.Get("StorageServiceCode").(string)

	ui.Say(fmt.Sprintf("Attaching system storage to the server ..."))

	args := protocol.BootDeviceStorageConnect{
		GisServiceCode: config.GisServiceCode,
		IvmServiceCode: ivm,
	}
	if strings.HasPrefix(storageServiceCode, "iba") {
		args.IbaServiceCode = storageServiceCode
	} else if strings.HasPrefix(storageServiceCode, "ica") {
		args.IcaServiceCode = storageServiceCode
	} else {
		ui.Error(fmt.Sprintf("Unknown storage type: %s", storageServiceCode))
		state.Put("error", fmt.Errorf("Unknown storage type: %s", storageServiceCode))
		return multistep.ActionHalt
	}
	resp := protocol.BootDeviceStorageConnectResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(fmt.Sprintf("Cannot attach system storage to the server. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if err := p2pubapi.WaitVM(api, config.GisServiceCode, ivm,
		p2pubapi.InService, p2pubapi.Stopped, 5*time.Minute); err != nil {
		ui.Error(fmt.Sprintf("Virtual server could not be ready. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepAttachSystemStorage) Cleanup(state multistep.StateBag) {
}
