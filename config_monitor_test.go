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

		addNamePatternList()
		time.Sleep(500 * time.Millisecond)

		config := cm.GetConfig()
		require.Greater(
			t,
			len(config),
			0)
	}
}

func prepareEmptyConfig() {
	configContent := "{}\n"
	_ = ioutil.WriteFile(TestConfigPath, []byte(configContent), 0644)
}

func addNamePatternList() {
	configContent := strings.TrimSpace(`
	{
		"namePatternList" : [
			"go test"
		]
	}
	`)
	_ = ioutil.WriteFile(TestConfigPath, []byte(configContent), 0644)
}
