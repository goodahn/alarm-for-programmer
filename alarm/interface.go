package alarm

import (
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
)

type Alarmer interface {
	GetMonitoringCommandList() []string
	GetMonitoringPeriod() time.Duration
	GetAlarmConfig() map[string]string
	GetAlarmCountMap() map[string]int
	GetTotalAlarmCountOfMonitoringCommand(monitoringCommand string) int

	Init(configMonitor *alarm.ConfigMonitor)

	Start()
	IsStarted() bool

	Stop()
}
