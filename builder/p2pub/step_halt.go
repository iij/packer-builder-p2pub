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

type stepHalt struct {
}

func (s *stepHalt) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Stopping server ..."))

	ivm := state.Get("IvmServiceCode").(string)

	args := protocol.VMPower{
		GisServiceCode: config.GisServiceCode,
		IvmServiceCode: ivm,
		Power:          "Off",
	}

	resp := protocol.VMPowerResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(fmt.Sprintf("Failed to stop server. reason: %s", err.Error()))
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

func (s *stepHalt) Cleanup(state multistep.StateBag) {

}
