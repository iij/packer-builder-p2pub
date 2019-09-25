package p2pub

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func getStandardPrivateNetworkIPAddress(resp protocol.VMAddResponse) (string, error) {
	address := ""
	for _, elm := range resp.NetworkList {
		if elm.NetworkType == "PrivateStandard" {
			for _, addr := range elm.IpAddressList {
				if addr.IPv4.IpAddress != "" {
					address = addr.IPv4.IpAddress
					break
				}
			}
			if address != "" {
				break
			}
		}
	}
	if address == "" {
		return "", fmt.Errorf("IP address is not assigned on Standard Private Network")
	}
	return address, nil
}

type stepAllocateIPAddress struct {
}

func (s *stepAllocateIPAddress) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Allocating IP address ..."))

	ivm := state.Get("IvmServiceCode").(string)
	createVMResponse := state.Get("CreateVMResponse").(protocol.VMAddResponse)

	if config.DisableGlobalAddress {
		address, err := getStandardPrivateNetworkIPAddress(createVMResponse)
		if err != nil {
			ui.Error("Failed to get private IP address to connect. ")
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("IpAddress", address)
	} else {
		args := protocol.GlobalAddressAllocate{
			GisServiceCode: config.GisServiceCode,
			IvmServiceCode: ivm,
		}
		resp := protocol.GlobalAddressAllocateResponse{}
		if err := p2pubapi.Call(*api, args, &resp); err != nil {
			ui.Error(fmt.Sprintf("Cannot allocate global ip address to the server. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("IpAddress", resp.IPv4.IpAddress)
	}

	return multistep.ActionContinue
}

func (s *stepAllocateIPAddress) Cleanup(state multistep.StateBag) {
}
