package alarm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlarmer(t *testing.T) {
	t.Run("AlarmCount", CheckAlarmCount())
}

func CheckAlarmCount() func(*testing.T) {
	return func(t *testing.T) {
		alarmer := NewAlarmer([]string{
			"bash test",
			"THERE WILL BE NO PROCESS LIKE THIS",
		})

		executeBashScript()

		require.Equal(t, 0, alarmer.GetAlarmCount("THERE WILL BE NO PROCESS LIKE THIS"))
		require.Greater(t, alarmer.GetAlarmCount("bash test"), 0)
	}
}

func executeBashScript() {
}
