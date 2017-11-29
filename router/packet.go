package router

import (
	"strings"
	"strconv"
)

const (
	DATA = 0
	ACK = 1
	NACK = 2
	SYN = 3
	END = 4
	)

type Packet struct{
	Type uint8
	SequenceNumber uint32

	Address []byte
	Port uint16

	Data []byte

	RouterConnection *RouterConnection

}

func hostToBytes(host string) []byte{
	nums := strings.Split(host, ".")
	bytes := make([]byte, 4)

	for i:=0; i < 4; i++{
		numAsInt, _ := strconv.Atoi(nums[i])
		bytes[i] = uint8(numAsInt)
	}

	return bytes
}
