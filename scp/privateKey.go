package scp

import (
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/errors"
)

// ParsePrivateKey is
func ParsePrivateKey(path string) (ssh.AuthMethod, error) {
	expanded := os.ExpandEnv(path)

	keyContent, err := ioutil.ReadFile(expanded)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key file %s", expanded)
	}

	key, err := ssh.ParsePrivateKey(keyContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key file contents")
	}

	return ssh.PublicKeys(key), nil
}
