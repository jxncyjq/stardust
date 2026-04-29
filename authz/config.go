package authz

import (
	"fmt"
	"strings"
)

const (
	// DefaultAdapter is the default Casbin policy adapter.
	DefaultAdapter = "file"
	// GormAdapter enables policy storage in a SQL database via Casbin GORM adapter.
	GormAdapter = "gorm"
	// DefaultSubjectKey is the default request subject key.
	DefaultSubjectKey = "id"
	// DefaultObjectMode is the default request object extraction mode.
	DefaultObjectMode = "route"
	// PathObjectMode extracts the authorization object from the request path.
	PathObjectMode = "path"
	// DefaultActionMode is the default request action extraction mode.
	DefaultActionMode = "method"
)

// Config configures the authorization engine.
type Config struct {
	ModelPath  string `json:"model_path"`
	PolicyPath string `json:"policy_path"`
	Adapter    string `json:"adapter"`
	// DatabaseName selects a named database from the databases component when Adapter is gorm.
	DatabaseName string `json:"database_name"`
	// Driver and DSN allow using gorm adapter directly without the databases component.
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`

	SubjectKey string `json:"subject_key"`
	ObjectMode string `json:"object_mode"`
	ActionMode string `json:"action_mode"`
}

// SetDefaults fills empty config fields with package defaults.
func (c *Config) SetDefaults() {
	if c.Adapter == "" {
		c.Adapter = DefaultAdapter
	}
	if c.SubjectKey == "" {
		c.SubjectKey = DefaultSubjectKey
	}
	if c.ObjectMode == "" {
		c.ObjectMode = DefaultObjectMode
	}
	if c.ActionMode == "" {
		c.ActionMode = DefaultActionMode
	}
}

// Validate checks whether the config can initialize the authz package.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("%w: config is nil", ErrInvalidConfig)
	}
	c.SetDefaults()
	c.Adapter = strings.ToLower(strings.TrimSpace(c.Adapter))
	c.SubjectKey = strings.TrimSpace(c.SubjectKey)
	c.ObjectMode = strings.ToLower(strings.TrimSpace(c.ObjectMode))
	c.ActionMode = strings.ToLower(strings.TrimSpace(c.ActionMode))

	if c.Adapter != DefaultAdapter {
		if c.Adapter != GormAdapter {
			return fmt.Errorf("%w: %s", ErrUnsupportedAdapter, c.Adapter)
		}
	}
	if strings.TrimSpace(c.ModelPath) == "" {
		return fmt.Errorf("%w: model_path is required", ErrInvalidConfig)
	}

	switch c.Adapter {
	case DefaultAdapter:
		if strings.TrimSpace(c.PolicyPath) == "" {
			return fmt.Errorf("%w: policy_path is required", ErrInvalidConfig)
		}
	case GormAdapter:
		if strings.TrimSpace(c.DatabaseName) == "" &&
			(strings.TrimSpace(c.Driver) == "" || strings.TrimSpace(c.DSN) == "") {
			return fmt.Errorf("%w: database_name or driver+dsn is required for gorm adapter", ErrInvalidConfig)
		}
	}
	if c.SubjectKey == "" {
		return fmt.Errorf("%w: subject_key is required", ErrInvalidConfig)
	}
	if c.ObjectMode != DefaultObjectMode && c.ObjectMode != PathObjectMode {
		return fmt.Errorf("%w: unsupported object_mode %q", ErrInvalidConfig, c.ObjectMode)
	}
	if c.ActionMode != DefaultActionMode {
		return fmt.Errorf("%w: unsupported action_mode %q", ErrInvalidConfig, c.ActionMode)
	}
	return nil
}
