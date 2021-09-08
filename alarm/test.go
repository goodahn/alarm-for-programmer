package alarm

import (
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func CheckAlarmCount(configPath string) func(*testing.T) {
	return func(t *testing.T) {
		alarmer := NewAlarmer(configPath)
		defer alarmer.Stop()
		require.True(t, alarmer.IsStarted())

		count := 5
		executeBashScriptManyTime(count)
		time.Sleep(time.Second)

		require.Equal(t, count, alarmer.GetTotalAlarmCountOfMonitoringCommand("bash test"))
		require.Equal(t, 0, alarmer.GetTotalAlarmCountOfMonitoringCommand("THERE WILL BE NO PROCESS LIKE THIS"))
	}
}

func CheckGetMonitoringCommandList(configPath string) func(*testing.T) {
	return func(t *testing.T) {
		alarmer := NewAlarmer(configPath)
		defer alarmer.Stop()
		require.True(t, alarmer.IsStarted())

		require.Equal(
			t,
			[]string{
				"bash test",
				"THERE WILL BE NO PROCESS LIKE THIS",
			},
			alarmer.GetMonitoringCommandList(),
		)
	}
}

func CheckGetMonitoringPeriod(configPath string) func(*testing.T) {
	return func(t *testing.T) {
		alarmer := NewAlarmer(configPath)
		defer alarmer.Stop()
		require.True(t, alarmer.IsStarted())

		require.Equal(
			t,
			time.Duration(50),
			alarmer.GetMonitoringPeriod(),
		)
	}
}

func CheckGetAlarmConfig(configPath string) func(*testing.T) {
	return func(t *testing.T) {
		alarmer := NewAlarmer(configPath)
		defer alarmer.Stop()
		require.True(t, alarmer.IsStarted())

		require.Equal(
			t,
			map[string]string{
				"type":           "slack-webhook",
				"webHookUrl":     "localhost",
				"requestTimeout": "10",
			},
			alarmer.GetAlarmConfig(),
		)
	}
}

func executeBashScriptManyTime(count int) {
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := exec.Command("bash", "test_monitoring_command.sh")
			c.Run()
		}()
	}
	wg.Wait()
}
