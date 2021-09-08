package alarm

import "testing"

func TestSlackWebhookAlarmer(t *testing.T) {
	t.Run("AlarmCount", CheckAlarmCount("test_config_for_slack_webhook_alarmer.json"))
}
