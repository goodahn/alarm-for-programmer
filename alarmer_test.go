package alarm

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAlarmer(t *testing.T) {
	t.Run("AlarmCount", CheckAlarmCount())
}

func CheckAlarmCount() func(*testing.T) {
	return func(t *testing.T) {
		prepareConfigJson()
		alarmer := NewAlarmer("test_config.json")
		defer alarmer.Stop()

		count := 5
		executeBashScriptManyTime(count)
		time.Sleep(3 * defaultPeriod)

		require.Equal(t, count, alarmer.GetTotalAlarmCountOfNamePattern("bash test"))
		require.Equal(t, 0, alarmer.GetTotalAlarmCountOfNamePattern("THERE WILL BE NO PROCESS LIKE THIS"))
	}
}

func prepareConfigJson() {
	configContent := strings.TrimSpace(`
	{
		"namePatternList" : [
			"bash test",
			"THERE WILL BE NO PROCESS LIKE THIS"
		],
		"alarmConfig" : {
			"type": "slack-webhook",
			"webHookUrl":"localhost"
		}
	}
	`)
	_ = ioutil.WriteFile(TestConfigPath, []byte(configContent), 0644)
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
