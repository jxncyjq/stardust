package logs

import (
	"encoding/json"
	"testing"
)

func TestLogger(t *testing.T) {
	configMap := map[string]interface{}{
		"global": map[string]interface{}{
			"appName":    "testApp",
			"appVersion": "1.0.0",
		}}
	conf, err := json.Marshal(configMap)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	Init(conf)

	instanceLogger.Info("This is an info message")
	instanceLogger.Warn("This is a warning message")
	instanceLogger.Error("This is an error message")
}
