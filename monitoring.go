package alarm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
}

type ProcessInfo struct {
	uid    string
	cmd    string
	pid    int
	status int
}

type Monitor struct {
	totalAlarmCount          int
	monitoringPeriod         time.Duration
	webHookURL               string
	cmdList                  []string
	processInfoMap           map[string]*ProcessInfo
	mutexForProcessStatusMap *sync.Mutex

	stop context.CancelFunc
}

func NewCommandMonitor(configPath string) (m Monitor) {
	m = Monitor{
		monitoringPeriod: time.Second,
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
	fmt.Println(config)

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
	ctx, done := context.WithCancel(context.Background())
	m.stop = done
	isStopped := false

	// go routine for getting stop call
	go func() {
		select {
		case <-ctx.Done():
			isStopped = true
		}
	}()

	go func() {
		for {
			if isStopped {
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
	m.stop()
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
			pid:    p.pid,
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
					uid := makeUID(cmd, pid)
					updateProcessInfoMap(processInfoMap,
						mutexForProcessStatusMap,
						uid,
						cmd,
						pid,
						Executing)
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
						_p.pid,
						Executing)
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
						p.pid,
						Finished)
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
	uid, cmd string, pid, status int) {
	mutex.Lock()
	defer mutex.Unlock()
	pim[uid] = &ProcessInfo{
		uid:    uid,
		cmd:    cmd,
		pid:    pid,
		status: status,
	}
}

func deleteProcessInfo(pim map[string]*ProcessInfo, mutex *sync.Mutex, uid string) {
	mutex.Lock()
	delete(pim, uid)
	mutex.Unlock()
}
