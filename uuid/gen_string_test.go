package uuid

import (
	"fmt"
	"testing"
)

func TestGenString(t *testing.T) {
	result := GenString(8)
	if len(result) != 8 {
		fmt.Println(result)
		t.Errorf("expected string length 8, got %d", len(result))
	}
	fmt.Print(result)
	t.Log(result)

}

func TestGenDateRnString(t *testing.T) {
	result := GenDateRnString("01")
	fmt.Println(result)
	t.Log(result)
}

func TestUUid(t *testing.T) {
	uuid := GetUuidString()
	fmt.Println(uuid)
	t.Log(uuid)
}
