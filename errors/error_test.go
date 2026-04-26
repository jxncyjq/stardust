package errors

import (
	"fmt"
	"strconv"
	"testing"
)

func TestErrorInfo(t *testing.T) {
	err1 := New("test1", 0)
	err2 := New("test2", 0)
	err := WithStack(err1, 0, err2)
	var v1 = "a"
	i, err := strconv.Atoi(v1)
	if err == nil {
		t.Log(i)
	}
	err1.extraErrs = append(err1.extraErrs, err)
	s := fmt.Sprintf("%s:%s", err1.ToStr(), err1.ToStrByExtra())
	t.Log(s)
}
