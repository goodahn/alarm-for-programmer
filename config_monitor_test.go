package alarm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigMonitor(t *testing.T) {
	t.Run("EmptyConfig", CheckEmptyConfig())
	t.Run("ChangeConfig", CheckChangeConfig())
}

func CheckEmptyConfig() func(*testing.T) {
	return func(t *testing.T) {
		prepareEmptyConfig()
		cm := NewConfigMonitor("test_config.json")

		require.Equal(
			t,
			0,
			len(cm.GetConfig()))
	}
}

func CheckChangeConfig() func(*testing.T) {
	return func(t *testing.T) {
		prepareEmptyConfig()
		cm := NewConfigMonitor("test_config.json")

		require.Equal(
			t,
			0,
			len(cm.GetConfig()))

		addTargetProcessList()
		require.Equal(
			t,
			1,
			len(cm.GetConfig()))
	}
}

func prepareEmptyConfig() {
}

func addTargetProcessList() {
}
