package alarm

import "time"

const (
	NotFoundInNamePatternList = "NOT_FOUND_IN_TARGET_PROCESS_LIST"
	ProcessNeverStarted       = "PROCESS_NEVER_STARTED"
	ProcessStarted            = "PROCESS_STARTED"
	ProcessFinished           = "PROCESS_FINISHED"
)

var (
	CheckProcessStatus = map[string]bool{
		NotFoundInNamePatternList: true,
		ProcessNeverStarted:       true,
		ProcessStarted:            true,
		ProcessFinished:           true,
	}
)

type ProcessStatus struct {
	pid       int
	status    string
	timestamp time.Time
}

func NewProcessStatus(pid int, status string) ProcessStatus {
	ps := ProcessStatus{}
	ps.Init(pid, status)
	return ps
}

func (ps *ProcessStatus) Init(pid int, status string) {
	_, ok := CheckProcessStatus[status]
	if ok == false {
		return
	}
	ps.pid = pid
	ps.status = status
	ps.timestamp = time.Now()
}

func (ps *ProcessStatus) Pid() int {
	return ps.pid
}

func (ps *ProcessStatus) Status() string {
	return ps.status
}

func (ps *ProcessStatus) TimeStamp() time.Time {
	return ps.timestamp
}

func (ps *ProcessStatus) TimeStampInString() string {
	return ps.timestamp.String()
}

func (ps *ProcessStatus) SetTimestamp(timestamp time.Time) {
	ps.timestamp = timestamp
}
