package scp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

var (
	currentDepth int
	lastDepth    int
)

// UploadDirectory is used to transfer a directory from the local system to a remote system
// over an open SSH connection
func UploadDirectory(stdin io.WriteCloser, path string, preserveTimes bool) error {
	err := filepath.Walk(path, func(path string, fileInfo os.FileInfo, err error) error {
		currentDepth := len(filepath.SplitList(path))
		fmt.Printf("currentDepth: %d, lastDepth: %d\n", currentDepth, lastDepth)
		if currentDepth < lastDepth {
			fmt.Println("E")
			_, err = stdin.Write([]byte("E\n"))
			if err != nil {
				return errors.Wrap(err, "failed to terminate the connection")
			}
		}

		if err != nil {
			return errors.Wrap(err, "error occurred while walking directory")
		}

		if !fileInfo.IsDir() {
			lastDepth = currentDepth
			return UploadFile(stdin, path, preserveTimes)
		}

		// D0644 12 test
		// would mean: 12 byte file named test which has 0644 for it's filesystem permissions
		beginMessage := fmt.Sprintf("D%04o 0 %s\n",
			fileInfo.Mode().Perm(),
			fileInfo.Name())

		if preserveTimes {
			accessTimes := fmt.Sprintf("T%d 0 %d 0\n",
				fileInfo.ModTime().Unix(),
				time.Now().Unix())
			fmt.Print(accessTimes)
			_, err = stdin.Write([]byte(accessTimes))
			if err != nil {
				return errors.Wrap(err, "failed to begin transfer")
			}
		}

		fmt.Print(beginMessage)
		_, err = stdin.Write([]byte(beginMessage))
		if err != nil {
			return errors.Wrap(err, "failed to begin transfer")
		}

		lastDepth = currentDepth
		return nil
	})

	fmt.Println("E")
	_, err = stdin.Write([]byte("E\n"))
	if err != nil {
		return errors.Wrap(err, "failed to terminate the connection")
	}
	return err
}

// UploadFile is used to transfer a file from the local system to a remote system
// over an open SSH connection
func UploadFile(stdin io.WriteCloser, path string, preserveTimes bool) error {
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

	if preserveTimes {
		accessTimes := fmt.Sprintf("T%d 0 %d 0\n",
			fileInfo.ModTime().Unix(),
			time.Now().Unix())
		fmt.Print(accessTimes)
		_, err = stdin.Write([]byte(accessTimes))
		if err != nil {
			return errors.Wrap(err, "failed to begin transfer")
		}
	}

	fmt.Print(beginMessage)
	_, err = stdin.Write([]byte(beginMessage))
	if err != nil {
		return errors.Wrap(err, "failed to begin transfer")
	}

	fmt.Println("copying file")
	file.Seek(0, 0)
	_, err = io.Copy(stdin, file)
	if err != nil {
		return errors.Wrap(err, "failed to copy file contents")
	}

	fmt.Println("file transfer complete")
	_, err = stdin.Write([]byte("\x00"))
	if err != nil {
		return errors.Wrap(err, "failed to terminate the connection")
	}

	return nil
}
