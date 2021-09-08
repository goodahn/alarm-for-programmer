package alarm

import (
	"sync"
	"time"
)

type ProcessInfoMonitor struct {
	monitoringCommandList                   []string
	processStatusHistoryByMonitoringCommand map[string](map[int]([]ProcessStatus))
	monitoringPeriod                        time.Duration
	start                                   bool

	mutexForSynchronousMethodCall sync.Mutex
	mutexForProcessStatusHistory  sync.Mutex

	now time.Time

	processInfoReader *ProcessInfoReader
}

func NewProcessInfoMonitor(monitoringCommandList []string) *ProcessInfoMonitor {
	pim := &ProcessInfoMonitor{}
	pim.Init()
	pim.SetMonitoringCommandList(monitoringCommandList)
	pim.Start()
	time.Sleep(defaultPeriod)
	return pim
}

func (pim *ProcessInfoMonitor) Init() {
	pim.processStatusHistoryByMonitoringCommand = map[string](map[int]([]ProcessStatus)){}
	pim.processInfoReader = NewProcessInfoReader()
	pim.SetPeriod(defaultPeriod)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (pim *ProcessInfoMonitor) Start() {
	pim.mutexForSynchronousMethodCall.Lock()
	defer pim.mutexForSynchronousMethodCall.Unlock()

	if pim.start {
		return
	}
	pim.start = true

	pim.processInfoReader.Start()
	go func() {
		for {
			if !pim.start {
				break
			}

			pim.now = time.Now()
			pim.updateTargetProcessStatusHistory()
		}
	}()
}

func (pim *ProcessInfoMonitor) updateTargetProcessStatusHistory() {
	for _, namePattern := range pim.monitoringCommandList {
		pim.updateProcessStatusHistoryByMonitoringCommand(namePattern)
	}
}

func (pim *ProcessInfoMonitor) updateProcessStatusHistoryByMonitoringCommand(namePattern string) {
	changedProcessStatusMap := pim.getChangedProcessStatus(namePattern)
	if len(changedProcessStatusMap) == 0 {
		return
	}
	pim.mutexForProcessStatusHistory.Lock()
	defer pim.mutexForProcessStatusHistory.Unlock()
	for pid, changedProcessStatus := range changedProcessStatusMap {
		pim.processStatusHistoryByMonitoringCommand[namePattern][pid] = append(
			pim.processStatusHistoryByMonitoringCommand[namePattern][pid],
			changedProcessStatus,
		)
	}
}

func (pim *ProcessInfoMonitor) getChangedProcessStatus(namePattern string) map[int]ProcessStatus {
	changedProcessStatusHistory := map[int]ProcessStatus{}

	pim.mutexForProcessStatusHistory.Lock()
	defer pim.mutexForProcessStatusHistory.Unlock()
	processStatusHistory := pim.processStatusHistoryByMonitoringCommand[namePattern]

	pidList := pim.processInfoReader.GetPidListByName(namePattern)
	for _, pid := range pidList {
		latestProcessStatus := pim.getUpdatedProcessStatusInLogWithTimestamp(processStatusHistory, pid)
		if (latestProcessStatus == ProcessStatus{}) {
			continue
		}
		changedProcessStatusHistory[pid] = latestProcessStatus
	}
	return changedProcessStatusHistory
}

func (pim *ProcessInfoMonitor) getUpdatedProcessStatusInLogWithTimestamp(processStatusHistory map[int]([]ProcessStatus), pid int) ProcessStatus {
	isExecuting := pim.processInfoReader.IsExecuting(pid)
	if isExecuting {
		// process is started after its previous execution
		mostRecentProcessStatus := findProcessStatusInHistory(processStatusHistory, pid)
		if mostRecentProcessStatus.Status() == ProcessFinished ||
			mostRecentProcessStatus.Status() == ProcessNeverStarted {
			return NewProcessStatus(pid, ProcessStarted)
		}
	} else {
		// process is finshed after start
		mostRecentProcessStatus := findProcessStatusInHistory(processStatusHistory, pid)
		if mostRecentProcessStatus.Status() == ProcessStarted {
			return NewProcessStatus(pid, ProcessFinished)
		}
	}
	return ProcessStatus{}
}

func (pim *ProcessInfoMonitor) Stop() {
	pim.start = false
	pim.processInfoReader.Stop()
}

func (pim *ProcessInfoMonitor) SetPeriod(period time.Duration) {
	pim.monitoringPeriod = period
	pim.processInfoReader.SetPeriod(period)
}

func (pim *ProcessInfoMonitor) SetMonitoringCommandList(monitoringCommandList []string) {
	pim.monitoringCommandList = monitoringCommandList
	for _, namePattern := range monitoringCommandList {
		pim.processStatusHistoryByMonitoringCommand[namePattern] = map[int]([]ProcessStatus){}
	}
}

func (pim *ProcessInfoMonitor) GetProcessStatusLogByMonitoringCommand(namePattern string) map[int]([]ProcessStatus) {
	pim.mutexForProcessStatusHistory.Lock()
	defer pim.mutexForProcessStatusHistory.Unlock()
	return pim.getProcessStatusHistory(namePattern)
}

func (pim *ProcessInfoMonitor) getProcessStatusHistory(namePattern string) map[int]([]ProcessStatus) {
	processStatusHistory, ok := pim.processStatusHistoryByMonitoringCommand[namePattern]
	if !ok {
		return map[int]([]ProcessStatus){}
	}
	return processStatusHistory
}

// findProcessStatusInHistory finds the ProcessStatus
// whose timestamp is the biggest while its timestamp is smaller than given timestamp
func findProcessStatusInHistory(wholeProcessStatusHistory map[int]([]ProcessStatus), pid int) ProcessStatus {
	if len(wholeProcessStatusHistory) == 0 {
		return NewProcessStatus(pid, ProcessNeverStarted)
	}
	mostRecentProcessStatus := NewProcessStatus(pid, ProcessNeverStarted)
	mostRecentProcessStatus.SetTimestamp(time.Time{})

	for _pid, processStatusHistory := range wholeProcessStatusHistory {
		if pid != _pid {
			continue
		}
		mostRecentProcessStatus = processStatusHistory[len(processStatusHistory)-1]
	}
	return mostRecentProcessStatus
}
