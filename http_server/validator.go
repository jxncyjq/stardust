package httpServer

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator 自定义 Validator
type CustomValidator struct {
	Validator *validator.Validate
}

// Validate 实现 Validate 方法
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
