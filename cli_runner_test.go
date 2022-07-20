package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type runningCLI struct {
	output       io.Reader
	waitToFinish func() error
	cleanup      func() error
}

func startCLI(args []string, env map[string]string, workDir, binaryPath string) (*runningCLI, error) {

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", workDir),
	)

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	opr, opw := io.Pipe()

	cmd.Stdout = opw
	cmd.Stderr = opw

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("while starting cmd: %w", err)
	}

	processDoneChan := make(chan int)

	go func() {
		_, err := cmd.Process.Wait()
		if err != nil {
			fmt.Println(fmt.Errorf("while waiting for the process: %w", err))
			return
		}
		opw.Close()
		close(processDoneChan)
	}()

	return &runningCLI{
		output: opr,
		waitToFinish: func() error {
			select {
			case <-processDoneChan:
				// all good, server is down
			case <-time.NewTimer(3 * time.Second).C:
				return errors.New("timed out waiting for cmd to finish")
			}
			return nil
		},
		cleanup: func() error {
			select {
			case <-processDoneChan:
				// all good, server is down
			default:
				err = cmd.Process.Kill()
				if err != nil {
					return fmt.Errorf("while killing process: %w", err)
				}
				select {
				case <-time.NewTimer(3 * time.Second).C:
					return fmt.Errorf("timed out while shutting down server")
				case <-processDoneChan:
					// all good, server is down now, continue
				}
			}

			return nil

		},
	}, nil

}

func runCLI(args []string, env map[string]string, workDir, binaryPath string) (stdout, stderr string, err error) {

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", workDir),
	)

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stout := new(bytes.Buffer)
	sterr := new(bytes.Buffer)

	cmd.Stdout = stout
	cmd.Stderr = sterr

	err = cmd.Run()
	if err != nil {
		return stout.String(), sterr.String(), fmt.Errorf("while running cmd: %w", err)
	}

	return stout.String(), sterr.String(), nil

}
