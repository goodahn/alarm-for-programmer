package alarm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	cmd            string
	pid            int
	binaryLocation string
}

func (pi *ProcessInfo) Cmd() (cmd string) {
	return pi.cmd
}

func (pi *ProcessInfo) Pid() (pid int) {
	return pi.pid
}

func (pi *ProcessInfo) BinaryLocation() (binaryLocation string) {
	return pi.binaryLocation
}

func GetProcessInfoList() ([]ProcessInfo, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]ProcessInfo, 0, 50)
	for {
		names, err := d.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			if isProcessPid(name) {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.Atoi(name)
			if err != nil {
				continue
			}

			cmd, err := getCmdOfProcessByPid(pid)
			if err != nil {
				continue
			}

			binaryLocation, err := getBinaryLocation(pid)
			if err != nil {
				continue
			}

			p, err := newProcessInfo(cmd, pid, binaryLocation)
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func isProcessPid(name string) bool {
	return name[0] < '0' || name[0] > '9'
}

func getCmdOfProcessByPid(pid int) (string, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}
	cmd := strings.Join(strings.Split(string(bytes.TrimRight(data, string("\x00"))), string(byte(0))), " ")
	return cmd, nil
}

func getBinaryLocation(pid int) (string, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return "", err
	}
	environ := strings.Split(string(data), string(byte(0)))
	binaryLocation := ""
	for _, env := range environ {
		if strings.Contains(env, "PWD") {
			binaryLocation = strings.Split(env, "=")[1]
		}
	}
	return binaryLocation, nil
}

func newProcessInfo(cmd string, pid int, binaryLocation string) (newProcessInfo ProcessInfo, err error) {
	newProcessInfo = ProcessInfo{
		cmd:            cmd,
		pid:            pid,
		binaryLocation: binaryLocation,
	}
	return newProcessInfo, nil
}
