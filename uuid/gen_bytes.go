package uuid

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

func GenBytes(len int) []byte {
	var container []byte = make([]byte, len)
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container = append(container, str[randomInt.Int64()])
	}
	return container
}
