package alarm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	// constants for process status
	Unstarted = 0
	Executing = 1
	Finished  = 2

	// constants for checkExecutingOrNot
	notExecuted = 0
	executing   = 1
)

type Process interface {
	Cmd() string
	Pid() int
	Cwd() string
}

type ProcessInfo struct {
	uid     string
	cmd     string
	cwd     string
	name    string
	pid     int
	status  int
	succeed bool
}

type Monitor struct {
	totalAlarmCount          int
	monitoringPeriod         time.Duration
	webHookURL               string
	cmdList                  []string
	processInfoMap           map[string]*ProcessInfo
	mutexForProcessStatusMap *sync.Mutex
	isStopped                bool
}

func NewCommandMonitor(configPath string) (m Monitor) {
	m = Monitor{
		monitoringPeriod: time.Second,
		isStopped:        false,
	}
	m.processInfoMap = map[string]*ProcessInfo{}
	m.mutexForProcessStatusMap = &sync.Mutex{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		errMsg := fmt.Sprintf("error occured during reading config file: %v", err)
		panic(errMsg)
	}
	config := map[string]interface{}{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		errMsg := fmt.Sprintf("error occured during parsing config file: %v", err)
		panic(errMsg)
	}

	if _url, ok := config["slackWebHookURL"]; ok {
		if url, ok := _url.(string); ok {
			m.setWebHookURL(url)
		}
	}
	if rawCmdList, ok := config["commandList"]; ok {
		if _cmdList, ok := rawCmdList.([]interface{}); ok {
			cmdList := []string{}
			for _, _cmd := range _cmdList {
				if cmd, ok := _cmd.(string); ok {
					cmdList = append(cmdList, cmd)
				}
			}
			m.setCommandList(cmdList)
		}
	}
	return m
}

func (m *Monitor) Start() {
	go func() {
		for {
			if m.isStopped {
				break
			}

			err := m.updateProcessesStatus()
			if err != nil {
				log.Printf("update processes status failed: %v\n", err)
				return
			}

			err = m.alarm()
			if err != nil {
				log.Printf("alarm failed: %v\n", err)
			}
			time.Sleep(m.monitoringPeriod)
		}
	}()
}

func (m *Monitor) Stop() {
	m.isStopped = true
}

func (m *Monitor) SetMonitoringPeriod(period time.Duration) {
	m.monitoringPeriod = period
}

func (m *Monitor) TotalAlarmCount() (alarmCount int) {
	return m.totalAlarmCount
}

func (m *Monitor) WebHookURL() (url string) {
	return m.webHookURL
}

func (m *Monitor) CommandList() (cmdList []string) {
	return m.cmdList
}

func (m *Monitor) alarm() (err error) {
	for _, pi := range m.processInfoMap {
		if pi.status == Finished {
			var msg string
			// add package name
			if packageName := m.GetGoPackageName(pi.pid); packageName != "" {
				msg = fmt.Sprintf("[%s] %s (%d) is", packageName, pi.cmd, pi.pid)
			} else {
				msg = fmt.Sprintf("%s (%d) is", pi.cmd, pi.pid)
			}

			// add succeed or failed
			if pi.succeed {
				msg = fmt.Sprintf("%s succeed!", msg)
			} else {
				msg = fmt.Sprintf("%s failed!", msg)
			}

			log.Println(msg)
			if len(m.webHookURL) > 0 {
				data := map[string]string{
					"text": msg,
				}
				rawData, _ := json.Marshal(data)
				buff := bytes.NewBuffer(rawData)
				_, err := http.Post(m.webHookURL, "application/json", buff)
				if err != nil {
					log.Printf("web hook request is failed: %v\n", err)
				}
			}
			m.totalAlarmCount++
		}
	}
	return err
}

func (m *Monitor) updateProcessesStatus() (err error) {
	processList, err := getProcessList()
	if err != nil {
		return err
	}

	processInfoMap := map[string]*ProcessInfo{}
	mutexForProcessStatusMap := &sync.Mutex{}
	for uid, p := range m.processInfoMap {
		processInfoMap[uid] = &ProcessInfo{
			cmd:    p.cmd,
			name:   p.name,
			pid:    p.pid,
			cwd:    p.cwd,
			status: Unstarted,
		}
	}

	// get process info
	wg := sync.WaitGroup{}
	for _, process := range processList {
		wg.Add(1)
		go func(process Process) {
			defer wg.Done()
			pCmd := process.Cmd()
			for _, cmd := range m.cmdList {
				if strings.Contains(pCmd, cmd) {
					pid := process.Pid()
					cwd := process.Cwd()
					uid := makeUID(cmd, pid)
					updateProcessInfoMap(processInfoMap,
						mutexForProcessStatusMap,
						uid,
						pCmd,
						cwd,
						pid,
						Executing,
						false)
					break
				}
			}
		}(process)
	}
	wg.Wait()

	// delete finished process status
	for uid, _p := range processInfoMap {
		wg.Add(1)
		go func(_p *ProcessInfo, uid string) {
			defer wg.Done()
			if p, ok := getProcessInfo(m.processInfoMap,
				m.mutexForProcessStatusMap,
				uid); ok {
				if p.status == Finished &&
					_p.status == Unstarted {
					deleteProcessInfo(m.processInfoMap,
						m.mutexForProcessStatusMap,
						uid)
				}
			}
		}(_p, uid)
	}
	wg.Wait()

	// update newly added process
	for uid, _p := range processInfoMap {
		wg.Add(1)
		go func(_p *ProcessInfo, uid string) {
			defer wg.Done()
			if _, ok := getProcessInfo(m.processInfoMap,
				m.mutexForProcessStatusMap,
				uid); !ok {
				if _p.status == Executing {
					updateProcessInfoMap(m.processInfoMap,
						m.mutexForProcessStatusMap,
						_p.uid,
						_p.cmd,
						_p.cwd,
						_p.pid,
						Executing,
						_p.succeed)
					go m.WaitExitStatus(_p.pid)
				}
			}
		}(_p, uid)
	}
	wg.Wait()

	// update executing processes to finished
	for uid, _p := range processInfoMap {
		wg.Add(1)
		go func(_p *ProcessInfo, uid string) {
			defer wg.Done()
			if p, ok := getProcessInfo(m.processInfoMap,
				m.mutexForProcessStatusMap,
				uid); ok {
				if p.status == Executing &&
					_p.status == Unstarted {
					updateProcessInfoMap(m.processInfoMap,
						m.mutexForProcessStatusMap,
						p.uid,
						p.cmd,
						p.cwd,
						p.pid,
						Finished,
						p.succeed)
				}
			}
		}(_p, uid)
	}
	wg.Wait()
	return nil
}

func checkExecutingOrNot(cmd string) (status int) {
	return status
}

func (m *Monitor) setCommandList(cmdList []string) {
	m.cmdList = cmdList
}

func (m *Monitor) setWebHookURL(url string) {
	m.webHookURL = url
}

func (m *Monitor) GetPids(command string) (pids []int) {
	m.mutexForProcessStatusMap.Lock()
	defer m.mutexForProcessStatusMap.Unlock()

	pids = []int{}

	for _, pi := range m.processInfoMap {
		if strings.Contains(pi.cmd, command) {
			pids = append(pids, pi.pid)
		}
	}
	return pids
}

func (m *Monitor) GetCwd(pid int) (cwd string) {
	m.mutexForProcessStatusMap.Lock()
	defer m.mutexForProcessStatusMap.Unlock()

	for _, pi := range m.processInfoMap {
		if pi.pid == pid {
			cwd = pi.cwd
			break
		}
	}

	return cwd
}

func (m *Monitor) GetGoPackageName(pid int) (packageName string) {
	cwd := m.GetCwd(pid)
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fo, err := os.Open(cwd + "/" + f.Name())
		if err != nil {
			continue
		}

		reader := bufio.NewReader(fo)
		_line, _, err := reader.ReadLine()
		line := string(_line)
		if err == nil {
			if strings.Contains(line, "package") &&
				strings.Contains(line, "_test") {
				packageName = strings.Split(line, " ")[1]
				break
			}
		}
		fo.Close()
	}
	return packageName
}

func (m *Monitor) WaitExitStatus(pid int) (succeed bool) {
	c := exec.Command("wait", fmt.Sprint(pid))
	err := c.Run()
	m.mutexForProcessStatusMap.Lock()
	defer m.mutexForProcessStatusMap.Unlock()
	if _, ok := err.(*exec.ExitError); !ok {
		fmt.Println(pid, "succeed!")
		for _, pi := range m.processInfoMap {
			if pi.pid == pid {
				m.processInfoMap[pi.uid].succeed = true
				break
			}
		}
	} else {
		fmt.Println(pid, "failed!")
	}
	return succeed
}

func makeUID(cmd string, pid int) (uid string) {
	return fmt.Sprintf("%s||||%d", cmd, pid)
}

func getProcessInfo(pim map[string]*ProcessInfo, mutex *sync.Mutex,
	uid string) (p *ProcessInfo, ok bool) {
	mutex.Lock()
	defer mutex.Unlock()
	p, ok = pim[uid]
	return p, ok
}

func updateProcessInfoMap(pim map[string]*ProcessInfo, mutex *sync.Mutex,
	uid, cmd, cwd string, pid, status int, succeed bool) {
	mutex.Lock()
	defer mutex.Unlock()
	pim[uid] = &ProcessInfo{
		uid:     uid,
		cmd:     cmd,
		cwd:     cwd,
		pid:     pid,
		status:  status,
		succeed: succeed,
	}
}

func deleteProcessInfo(pim map[string]*ProcessInfo, mutex *sync.Mutex, uid string) {
	mutex.Lock()
	delete(pim, uid)
	mutex.Unlock()
}
