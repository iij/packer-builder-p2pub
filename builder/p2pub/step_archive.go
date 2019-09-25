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

// api
func waitForImage(api *p2pubapi.API, gis, iar, imageID string, timeout time.Duration) error {
	start := time.Now()
	for {
		arg := protocol.CustomOSImageGet{
			GisServiceCode: gis,
			IarServiceCode: iar,
			ImageId:        imageID,
		}
		res := protocol.CustomOSImageGetResponse{}
		if err := p2pubapi.Call(*api, arg, &res); err != nil {
			return err
		}
		if res.Status == "Created" {
			break
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout")
		}
		time.Sleep(20 * time.Second)
	}
	return nil
}

type stepArchive struct {
}

func (s *stepArchive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*p2pubapi.API)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Saving image ..."))

	serviceCode := state.Get("StorageServiceCode").(string)
	iar := state.Get("IarServiceCode").(string)

	if strings.HasPrefix(serviceCode, "iba") {
		args := protocol.CustomOSImageCreate{
			GisServiceCode: config.GisServiceCode,
			IarServiceCode: iar,
			IbaServiceCode: serviceCode,
			Name:           config.Label,
		}
		resp := protocol.CustomOSImageCreateResponse{}
		if err := p2pubapi.Call(*api, args, &resp); err != nil {
			ui.Error(fmt.Sprintf("Failed to save image. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}

		state.Put("ImageId", resp.ImageId)

	} else {

		args := protocol.OnlineBackup{
			GisServiceCode:     config.GisServiceCode,
			StorageServiceCode: serviceCode,
			Label:              config.Label,
		}
		resp := protocol.OnlineBackupResponse{}
		if err := p2pubapi.Call(*api, args, &resp); err != nil {
			ui.Error(fmt.Sprintf("Failed to save image. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}

		listArgs := protocol.CustomOSImageListGet{
			GisServiceCode: config.GisServiceCode,
			IarServiceCode: iar,
		}
		listResp := protocol.CustomOSImageListGetResponse{}
		if err := p2pubapi.Call(*api, listArgs, &listResp); err != nil {
			ui.Error(fmt.Sprintf("Failed to check saved image. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}

		// FIXME:
		for _, image := range listResp.ImageList {
			if image.Label == config.Label {
				state.Put("ImageId", image.ImageId)
				break
			}
		}

		if _, ok := state.GetOk("ImageId"); !ok {
			ui.Error(fmt.Sprintf("Failed to check saved image"))
			state.Put("error", fmt.Errorf("Failed to check saved image"))
			return multistep.ActionHalt
		}
	}

	if err := p2pubapi.WaitSystemStorage(api, config.GisServiceCode, serviceCode,
		p2pubapi.InService, p2pubapi.Attached, 30*time.Minute); err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if err := waitForImage(api, config.GisServiceCode, iar, state.Get("ImageId").(string), 30*time.Minute); err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Image saved. ID = %s:%s", iar, state.Get("ImageId").(string)))

	return multistep.ActionContinue
}

func (s *stepArchive) Cleanup(state multistep.StateBag) {
}
