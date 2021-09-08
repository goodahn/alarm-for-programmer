package alarm

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	TestConfigPath = "test_config.json"
)

func TestConfigMonitor(t *testing.T) {
	t.Run("EmptyConfig", CheckEmptyConfig())
	t.Run("ChangeConfig", CheckChangeConfig())
	t.Run("GetMonitoringCommandList", CheckGetMonitoringCommandList())
	t.Run("GetMonitoringPeriod", CheckGetMonitoringPeriod())
	t.Run("GetAlarmConfig", CheckGetAlarmConfig())
}

func CheckEmptyConfig() func(*testing.T) {
	return func(t *testing.T) {
		prepareEmptyConfig()
		cm := NewConfigMonitor(TestConfigPath)
		defer cm.Stop()

		require.Equal(
			t,
			0,
			len(cm.GetConfig()))
	}
}

func CheckChangeConfig() func(*testing.T) {
	return func(t *testing.T) {
		prepareEmptyConfig()
		cm := NewConfigMonitor(TestConfigPath)
		defer cm.Stop()

		require.Equal(
			t,
			0,
			len(cm.GetConfig()))

		addMonitoringCommandList()
		time.Sleep(500 * time.Millisecond)

		config := cm.GetConfig()
		require.Greater(
			t,
			len(config),
			0)
	}
}

func CheckGetMonitoringCommandList() func(*testing.T) {
	return func(t *testing.T) {
		cm := NewConfigMonitor(TestConfigPath)
		defer cm.Stop()

		require.Equal(
			t,
			[]string{
				"go test",
			},
			cm.GetMonitoringCommandList(),
		)
	}
}

func CheckGetMonitoringPeriod() func(*testing.T) {
	return func(t *testing.T) {
		cm := NewConfigMonitor(TestConfigPath)
		defer cm.Stop()

		require.Equal(
			t,
			1000,
			cm.GetMonitoringPeriod(),
		)
	}
}

func CheckGetAlarmConfig() func(*testing.T) {
	return func(t *testing.T) {
		cm := NewConfigMonitor(TestConfigPath)
		defer cm.Stop()

		require.Equal(
			t,
			map[string]string{
				"type":           "slack-webhook",
				"webHookUrl":     "localhost",
				"requestTimeout": "10",
			},
			cm.GetAlarmConfig(),
		)
	}
}

func prepareEmptyConfig() {
	configContent := "{}\n"
	_ = ioutil.WriteFile(TestConfigPath, []byte(configContent), 0644)
}

func addMonitoringCommandList() {
	configContent := strings.TrimSpace(`
	{
		"monitoringCommandList" : [
			"go test"
		],
		"monitoringPeriod": "1000",
		"alarmConfig" : {
			"type": "slack-webhook",
			"webHookUrl":"localhost",
			"requestTimeout": "10"
		}
	}
	`)
	_ = ioutil.WriteFile(TestConfigPath, []byte(configContent), 0644)
}
