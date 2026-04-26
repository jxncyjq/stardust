package errors

import (
	"fmt"
	"testing"
)

// Try 模拟 try/catch/finally 结构
func TestTryCatchChain(t *testing.T) {
	TryFunc(
		// Try block
		func() {
			fmt.Println("Running the try block.")
			// 触发 panic
			panic(New("Something went wrong.", 500))
		},
		// Catch block
		func(err interface{}) {
			fmt.Printf("Caught an error: %v\n", err)
		},
		// Finally block
		func() {
			fmt.Println("Finally block executed.")
		},
	)

	fmt.Println("Program continues...")

}
