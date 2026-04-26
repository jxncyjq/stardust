package errors

// Try 模拟 try/catch/finally 结构
func TryFunc(fun func(), catch func(err interface{}), finally func()) {
	defer func() {
		if finally != nil {
			finally()
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			if catch != nil {
				catch(r)
			} else {
				panic(r) // 无 catch 则重新抛出，不静默吞掉
			}
		}
	}()
	fun()
}
