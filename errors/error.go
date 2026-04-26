package errors

import "strings"

// StackError 错误自定义, 支持堆栈信息的错误结构体
type StackError struct {
	code      int     // 错误码
	err       error   // 错误信息 by error
	extraErrs []error // 额外的 error
}

// AppendErrors 追加额外的 error 信息
func (c *StackError) AppendErrors(extras ...error) *StackError {
	c.extraErrs = append(c.extraErrs, extras...)
	return c
}

// 实现系统自带的 error 接口
func (c *StackError) Error() string {
	return c.err.Error()
}

// Code 获取错误码. 只有设置了才会有.
func (c *StackError) Code() int {
	return c.code
}

// Msg 获取错误信息
func (c *StackError) Msg() string {
	if c.extraErrs != nil && len(c.extraErrs) > 0 {
		return c.ToStrByExtra()
	}
	return c.err.Error()
}

// Err 获取错误
func (c *StackError) Err() error {
	return c.err
}

// Errs 获取扩展的 error 列表
func (c *StackError) Errs() []error {
	return c.extraErrs
}

// ExistExtra 判断是否存在扩展错误信息
func (c *StackError) ExistExtra() bool {
	if len(c.extraErrs) == 0 {
		return false
	}
	return true
}

// ToStrByExtra 获取扩展的 error 拼接的字符串
func (c *StackError) ToStrByExtra() string {
	result := make([]string, 0, len(c.extraErrs)+1)
	result = append(result, c.err.Error())
	for _, err := range c.extraErrs {
		result = append(result, err.Error())
	}
	return strings.Join(result, "\n")
}

// ToStr 打印堆栈信息 by string, 带 \n 分隔
func (c *StackError) ToStr() string {
	return ToStr(c.err)
}

// ToArr 打印堆栈信息 by array
func (c *StackError) ToArr() []string {
	return ToArr(c.err)
}
