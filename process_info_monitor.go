package alarm

import "time"

const (
	NotFoundInTargetProcessList = "NOT_FOUND_IN_TARGET_PROCESS_LIST"
	ProcessStarted              = "PROCESS_STARTED"
	ProcessFinished             = "PROCESS_FINISHED"
)

type ProcessInfoMonitor struct {
	targetProcessList []string
	period            time.Duration
	stop              bool
}

func NewProcessInfoMonitor(targetProcessList []string) ProcessInfoMonitor {
	pim := ProcessInfoMonitor{}
	pim.SetTargetProcessList(targetProcessList)
	pim.SetPeriod(500 * time.Millisecond)
	pim.Start()
	time.Sleep(500 * time.Millisecond)
	return pim
}

func (pim *ProcessInfoMonitor) Start() {
}

func (pim *ProcessInfoMonitor) Stop() {
	pim.stop = true
}

func (pim *ProcessInfoMonitor) SetPeriod(period time.Duration) {
	pim.period = period
}

func (pim *ProcessInfoMonitor) SetTargetProcessList(targetProcessList []string) {
	pim.targetProcessList = targetProcessList
}

func (pim *ProcessInfoMonitor) GetCurrentProcessStatus(processName string) string {
	now := time.Now()
	return pim.getProcessStatusWithTimestamp(now)
}

func (pim *ProcessInfoMonitor) getProcessStatusWithTimestamp(timestamp time.Time) string {
	return ""
}
