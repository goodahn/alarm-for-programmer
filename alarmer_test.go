package alarm

import (
	"os/exec"
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
		alarmer := NewAlarmer([]string{
			"bash test",
			"THERE WILL BE NO PROCESS LIKE THIS",
		})
		defer alarmer.Stop()

		count := 5
		executeBashScriptManyTime(count)
		time.Sleep(2 * defaultPeriod)

		require.Equal(t, 0, alarmer.GetAlarmCount("THERE WILL BE NO PROCESS LIKE THIS"))
		require.Equal(t, count, alarmer.GetAlarmCount("bash test"))
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
