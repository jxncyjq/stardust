package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func resetConfStateForTest() {
	config = nil
	vp = nil
	initErr = nil
	initOnce = sync.Once{}
	AppName = ""
	AppVersion = ""
	RedisKeyPrefix = ""
}

func writeTempToml(content string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "testconf-*.toml")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		return "", err
	}
	_ = tmpfile.Close()
	return tmpfile.Name(), nil
}

func TestInitAndGet(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("runConfig")
		os.Unsetenv("ISDEBUG")
		os.Unsetenv("devConf")
		os.Unsetenv("prodConf")
		os.Unsetenv("GLOBAL_APP_NAME")
		resetConfStateForTest()
	})

	devContent := `[global]
app_name = "dev-app"
app_version = "0.0.1"
redis_key_prefix = "dev:key"

[logger]
level = 0
path = "./logs"

key1 = "dev_value"
`
	prodContent := `[global]
app_name = "prod-app"
app_version = "0.0.2"
redis_key_prefix = "prod:key"

[logger]
level = 1
path = "./prod-logs"

key1 = "prod_value"
`
	devPath, err := writeTempToml(devContent)
	if err != nil {
		t.Fatalf("Failed to create dev toml: %v", err)
	}
	defer os.Remove(devPath)
	prodPath, err := writeTempToml(prodContent)
	if err != nil {
		t.Fatalf("Failed to create prod toml: %v", err)
	}
	defer os.Remove(prodPath)

	os.Setenv("runConfig", devPath)
	resetConfStateForTest()
	Init()

	if got := GetAppName(); got != "dev-app" {
		t.Fatalf("APP_NAME = %s, want dev-app", got)
	}

	loggerBytes := Get("logger")
	if loggerBytes == nil {
		t.Fatal("Get(logger) = nil")
	}

	var logger map[string]interface{}
	if err := json.Unmarshal(loggerBytes, &logger); err != nil {
		t.Fatalf("json.Unmarshal(Get(logger)) error: %v", err)
	}
	if logger["path"] != "./logs" {
		t.Fatalf("logger.path = %v, want ./logs", logger["path"])
	}

	os.Setenv("runConfig", prodPath)
	Init()

	if got := GetAppName(); got != "dev-app" {
		t.Fatalf("APP_NAME = %s, want dev-app after second Init call", got)
	}
}

func TestInitWithDebugConfigAndEnvOverride(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("runConfig")
		os.Unsetenv("ISDEBUG")
		os.Unsetenv("devConf")
		os.Unsetenv("prodConf")
		os.Unsetenv("GLOBAL_APP_NAME")
		resetConfStateForTest()
	})

	devContent := `[global]
app_name = "dev-app"
app_version = "0.0.1"
redis_key_prefix = "dev:key"
`
	devPath, err := writeTempToml(devContent)
	if err != nil {
		t.Fatalf("Failed to create dev toml: %v", err)
	}
	defer os.Remove(devPath)

	os.Unsetenv("runConfig")
	os.Setenv("ISDEBUG", "1")
	os.Setenv("devConf", devPath)
	os.Unsetenv("prodConf")
	os.Setenv("GLOBAL_APP_NAME", "env-dev-app")

	resetConfStateForTest()
	Init()

	if got := GetAppName(); got != "env-dev-app" {
		t.Fatalf("APP_NAME = %s, want env-dev-app", got)
	}
}
