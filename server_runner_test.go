package main_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type runningServer struct {
	serverURL  string
	contentURL string
	shutdown   func() error
}

func startServer() (*runningServer, error) {
	td, err := os.MkdirTemp("", "kartusche-test")
	if err != nil {
		return nil, fmt.Errorf("while creating test temp dir: %w", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(td)
		}
	}()

	wd := filepath.Join(td, "work")
	err = os.Mkdir(wd, 0700)
	if err != nil {
		return nil, fmt.Errorf("while creating server work dir: %w", err)
	}

	cmd := exec.Command("go", "run", ".", "server")
	cmd.Env = append(
		os.Environ(),
		"CONTROLLER_ADDR=localhost:0",
		"KARTUSCHES_ADDR=localhost:0",
		fmt.Sprintf("WORK_DIR=%s", wd),
	)

	opr, opw := io.Pipe()
	output := new(bytes.Buffer)
	outWriter := io.MultiWriter(opw, output)
	cmd.Stdout = outWriter
	cmd.Stderr = outWriter

	logChan := make(chan map[string]string)
	processDoneChan := make(chan int)
	go parseLogs(opr, logChan)

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("while starting server: %w", err)
	}

	go func() {
		st, err := cmd.Process.Wait()
		if err != nil {
			// log failed wait?
			fmt.Println(fmt.Errorf("while waiting for the process: %w", err))
			return
		}
		processDoneChan <- st.ExitCode()
	}()

	timer := time.NewTimer(5 * time.Second)

	serverURL := ""
	contentURL := ""

	for serverURL == "" && contentURL == "" {
		select {
		case <-processDoneChan:
			return nil, fmt.Errorf("server has died before properly starting:\n%s\n", output.String())
		case <-timer.C:
			return nil, fmt.Errorf("server did not start within 5 seconds:\n%s\n", output.String())
		case log := <-logChan:

			switch log["msg"] {
			case "server started":
				serverURL = fmt.Sprintf("http://%s", log["addr"])
			case "listening for kartusche requests":
				contentURL = fmt.Sprintf("http://%s", log["addr"])
			}
		}
	}

	return &runningServer{
		serverURL:  serverURL,
		contentURL: contentURL,
		shutdown: func() error {
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

			return os.RemoveAll(td)
		},
	}, nil

}

func parseLogs(r io.Reader, lines chan map[string]string) {
	dec := json.NewDecoder(r)

	for {
		vals := map[string]string{}
		err := dec.Decode(&vals)
		if errors.Is(err, io.EOF) {
			return
		}
		lines <- vals
	}
}
