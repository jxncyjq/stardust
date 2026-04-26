package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceConf_Validate(t *testing.T) {
	conf := ServiceConf{
		Name: "user-service",
		Mode: "pro",
		Log:  LogConf{Level: "info"},
	}
	assert.NoError(t, conf.Validate())
}

func TestServiceConf_ValidateEmpty(t *testing.T) {
	conf := ServiceConf{}
	assert.Error(t, conf.Validate())
}

func TestServiceConf_SetUp(t *testing.T) {
	conf := ServiceConf{
		Name: "test-service",
		Mode: "dev",
		Log:  LogConf{Level: "debug"},
	}
	assert.NoError(t, conf.SetUp())
}
