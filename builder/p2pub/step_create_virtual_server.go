package p2pub

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

type stepCreateVirtualServer struct {
}

func (s *stepCreateVirtualServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Creating virtual server ..."))

	args := protocol.VMAdd{
		GisServiceCode: config.GisServiceCode,
		Type:           config.VMType,
		OSType:         state.Get("OSType").(string),
	}
	resp := protocol.VMAddResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(fmt.Sprintf("Failed to create virtual server. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Using virtual server '%s'", resp.ServiceCode))

	state.Put("IvmServiceCode", resp.ServiceCode)
	state.Put("CreateVMResponse", resp)

	return multistep.ActionContinue
}

func (s *stepCreateVirtualServer) Cleanup(state multistep.StateBag) {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say("Removing virtual server ...")

	ivm := state.Get("IvmServiceCode").(string)

	args := protocol.VMCancel{
		GisServiceCode: config.GisServiceCode,
		IvmServiceCode: ivm,
	}
	resp := protocol.VMCancelResponse{}
	if err := p2pubapi.Call(*api, args, &resp); err != nil {
		ui.Error(err.Error())
		ui.Say(fmt.Sprintf("Failed to remove virtual server '%s'. remove it manually.", ivm))
	}
}
