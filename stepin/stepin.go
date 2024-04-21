package stepin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func Exec(cmd string, args ...string) (string, error) {
	log.Println("Run command:", cmd, args)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, cmd, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func NewTmpFile(filename string, content []byte) (*os.File, DisposeFunc, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return nil, nil, err
	}

	if len(content) > 0 {
		n, err := tmpFile.Write(content)
		if err != nil {
			_ = os.Remove(tmpFile.Name())
			return nil, nil, err
		}
		if n != len([]byte(content)) {
			return nil, nil, fmt.Errorf("write %d bytes, but expect %d bytes", n, len([]byte(content)))
		}
	}

	return tmpFile,
		func() error {
			err = tmpFile.Close()
			if err != nil {
				_ = os.Remove(tmpFile.Name())
				return err
			}
			return os.Remove(tmpFile.Name())
		},
		nil
}
