package alarm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessInfoMonitor(t *testing.T) {
	t.Run("NamePatternList", CheckNamePatternList())
}

func CheckNamePatternList() func(*testing.T) {
	return func(t *testing.T) {
		pim := NewProcessInfoMonitor(
			[]string{"go test"},
		)
		defer pim.Stop()

		processStatusHistory := pim.GetCurrentProcessStatusHistoryByName("go test")
		for _, processStatus := range processStatusHistory {
			require.Equal(t,
				ProcessStarted,
				processStatus.Status())
		}

		processStatusHistory = pim.GetCurrentProcessStatusHistoryByName("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS")
		require.Zero(t, len(processStatusHistory))
	}
}
