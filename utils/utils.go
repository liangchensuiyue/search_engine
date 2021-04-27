package utils

import (
	"bytes"
	"encoding/binary"

	"github.com/spaolacci/murmur3" //哈希计算
)

var greate uint64 = 600000

//基础哈希计算
func CaculateWordId(word string) int32 {

	data := []byte(word)
	X := []byte{1}
	hasher := murmur3.New128() //128 ，2^128
	hasher.Write(data)
	v1, v2 := hasher.Sum128() //返回两个整数
	hasher.Write(X)
	v3, v4 := hasher.Sum128() //返回两个整数

	return int32((v1 + v2 + v3 + v4) % greate)
}
func IntTobytes(n int64) []byte {
	data := n
	bytebuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytebuffer, binary.BigEndian, data)
	return bytebuffer.Bytes()
}
func BytesToInt(bts []byte) int64 {
	bytebuffer := bytes.NewBuffer(bts)
	var data int64
	binary.Read(bytebuffer, binary.BigEndian, &data)

	return data
}
