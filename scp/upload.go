package scp

import (
	"fmt"
	"io"
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

	// C0644 12 test
	// would mean: 12 byte file named test which has 0644 for it's filesystem permissions
	beginMessage := fmt.Sprintf("C%04o %d %s\n",
		fileInfo.Mode(),
		fileInfo.Size(),
		fileInfo.Name())

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open file for transferring")
	}
	defer file.Close()

	readBuffer := make([]byte, 1)
	_, err = stdin.Write([]byte(beginMessage))
	if err != nil {
		return errors.Wrap(err, "failed to begin transfer")
	}
	stdout.Read(readBuffer)

	_, err = io.Copy(stdin, file)
	if err != nil {
		return errors.Wrap(err, "failed to copy file contents")
	}
	stdout.Read(readBuffer)

	_, err = stdin.Write([]byte("\x00"))
	if err != nil {
		return errors.Wrap(err, "failed to terminate the connection")
	}

	return nil
}
