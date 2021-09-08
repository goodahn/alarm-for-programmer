package alarm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessInfoMonitor(t *testing.T) {
	t.Run("MonitoringCommandList", CheckMonitoringCommandList())
}

func CheckMonitoringCommandList() func(*testing.T) {
	return func(t *testing.T) {
		pim := NewProcessInfoMonitor(
			[]string{"go test"},
		)
		defer pim.Stop()

		wholeProcessStatusHistory := pim.GetProcessStatusLogByMonitoringCommand("go test")
		for _, processStatusHistory := range wholeProcessStatusHistory {
			processStatus := processStatusHistory[len(processStatusHistory)-1]
			require.Equal(t,
				ProcessStarted,
				processStatus.Status())
		}

		wholeProcessStatusHistory = pim.GetProcessStatusLogByMonitoringCommand("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS")
		require.Zero(t, len(wholeProcessStatusHistory))
	}
}
