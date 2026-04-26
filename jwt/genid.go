package jwt

import (
	"bytes"
	"crypto/rand"
	"math/big"

	"github.com/bwmarrin/snowflake"
)

var (
	node, _ = snowflake.NewNode(1) // 1 is the machine ID, can be changed based on your setup
)

func GenId() string {
	return node.Generate().String()
}

func GenString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
