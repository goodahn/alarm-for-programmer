package alarm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessInfoMonitor(t *testing.T) {
	t.Run("TargetProcessList", CheckTargetProcessList())
}

func CheckTargetProcessList() func(*testing.T) {
	return func(t *testing.T) {
		pim := NewProcessInfoMonitor(
			[]string{"go test"},
		)

		ps := pim.GetCurrentProcessStatus("go test")
		require.Equal(t,
			ProcessStarted,
			ps.Status())

		ps = pim.GetCurrentProcessStatus("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS")
		require.Equal(t,
			NotFoundInTargetProcessList,
			ps.Status())
	}
}
