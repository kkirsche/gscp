package scp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// UploadFile is used to transfer a file from the local system to a remote system
// over an open SSH connection
func UploadFile(stdin io.WriteCloser, stdout io.Reader, path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve file information prior to file upload")
	}

	transferStartMessage := fmt.Sprintf("C%04o %d %s\n",
		fileInfo.Mode(),
		fileInfo.Size(),
		fileInfo.Name())

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to open file for transferring")
	}

	readBuffer := make([]byte, 1)
	stdin.Write([]byte(transferStartMessage))
	stdout.Read(readBuffer)
	stdin.Write(file)
	stdout.Read(readBuffer)
	stdin.Write([]byte("\x00"))

	return nil
}
