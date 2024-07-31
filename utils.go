package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		PrintError("IntToHex")
	}
	return buff.Bytes()
}

func PrintError(message string) {
	fmt.Println("Error:", message)
}
