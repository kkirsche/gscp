package scp

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHAgent is used to generate the SSH agent authentication method from the
// local host
func SSHAgent() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers), nil
}
