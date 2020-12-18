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

type UnixProcess struct {
	cmd string
	pid int
}

func (up *UnixProcess) Cmd() (cmd string) {
	return up.cmd
}

func (up *UnixProcess) Pid() (pid int) {
	return up.pid
}

func getProcessList() ([]Process, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]Process, 0, 50)
	for {
		names, err := d.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			// We only care if the name starts with a numeric
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				continue
			}
			cmd := strings.Join(strings.Split(string(bytes.TrimRight(data, string("\x00"))), string(byte(0))), " ")

			p, err := newUnixProcess(cmd, int(pid))
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func newUnixProcess(cmd string, pid int) (newUnixProcess *UnixProcess, err error) {
	newUnixProcess = &UnixProcess{
		cmd: cmd,
		pid: pid,
	}
	return newUnixProcess, nil
}
