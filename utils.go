package main

import (
	"github.com/brutella/can"
	"time"
)

func sendFrame(bus *can.Bus, id uint32, val [8]uint8) {
	frm := can.Frame{
		ID:     id,
		Length: uint8(len(val)),
		Flags:  0,
		Res0:   0,
		Res1:   0,
		Data:   val,
	}
	bus.Publish(frm)
}
func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
func lo8(x uint16) uint8 {
	return (uint8)((x) & 0xff)
}

func hi8(x uint16) uint8 {
	return (uint8)(((x) >> 8) & 0xff)
}
