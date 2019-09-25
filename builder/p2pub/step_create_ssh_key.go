package p2pub

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	// Using existing key.
	if config.Comm.SSHPrivateKeyFile != "" {
		ui.Say(fmt.Sprintf("Using existing SSH key: %s", config.Comm.SSHPrivateKeyFile))

		privateKeyBytes, err := config.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			ui.Error(fmt.Sprintf("Error reading ssh key. reason: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}

		config.Comm.SSHPrivateKey = privateKeyBytes
		config.Comm.SSHPublicKey = nil

		return multistep.ActionContinue
	}

	// Using temporary key and connect as "root"
	ui.Say(fmt.Sprintf("Creating temporary SSH key pair ..."))

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating temporary ssh key. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(priv),
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating temporary ssh key. reason: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}
	config.Comm.SSHPrivateKey = pem.EncodeToMemory(&privBlk)
	config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)
	config.RootSSHKey = string(config.Comm.SSHPublicKey)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
}
