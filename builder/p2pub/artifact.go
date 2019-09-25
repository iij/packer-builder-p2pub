package p2pub

import (
	"fmt"

	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

// Artifact ...
type Artifact struct {
	GisServiceCode string
	IarServiceCode string
	ImageID        string

	api *p2pubapi.API
}

// BuilderId ...
func (a Artifact) BuilderId() string {
	return "iij.p2pub"
}

// Files ...
func (a Artifact) Files() []string {
	return nil
}

// Id ...
func (a Artifact) Id() string {
	return a.IarServiceCode + ":" + a.ImageID
}

// String ...
func (a Artifact) String() string {
	return fmt.Sprintf("%s:%s (%s)", a.IarServiceCode, a.ImageID, a.GisServiceCode)
}

// State ...
func (a Artifact) State(name string) interface{} {
	return nil
}

// Destroy ...
func (a Artifact) Destroy() error {

	args := protocol.CustomOSImageDelete{
		GisServiceCode: a.GisServiceCode,
		IarServiceCode: a.IarServiceCode,
		ImageId:        a.ImageID,
	}
	resp := protocol.CustomOSImageDeleteResponse{}
	if err := p2pubapi.Call(*a.api, args, &resp); err != nil {
		return err
	}

	return nil
}
