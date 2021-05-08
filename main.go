package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brutella/can"
)

var lastSpeedValue = uint16(0)
var speedCounter = uint16(0xD0FF)
var seatbeltCounter = uint8(0)
var absBrakeCounter1 = uint8(0xF0)
var absFrame2 = uint8(0xB3)
var engineTemp2 = uint8(0x63)

func ignitionOn(bus *can.Bus) {
	sendFrame(bus,
		0x130,
		[8]uint8{0x45, 0x42, 0x69, 0x8f, 0xE2})
}

func ignitionStatus(bus *can.Bus) {
	sendFrame(bus, 0x26E, [8]uint8{0x40, 0x40, 0x7F, 0x50, 0xFF, 0xFF, 0xFF, 0xFF})
}

func sendSpeed(bus *can.Bus, speed uint16) {
	speedValue := speed + lastSpeedValue
	speedCounter = speedCounter & 0x0FFF
	speedCounter += 100 * 3
	sendFrame(bus, 0x1A6, [8]uint8{
		lo8(speedValue),
		hi8(speedValue),
		lo8(speedValue),
		hi8(speedValue),
		lo8(speedValue),
		hi8(speedValue),
		lo8(speedCounter),
		hi8(speedCounter),
	})
	lastSpeedValue = speedValue
}

func sendRPM(bus *can.Bus, rpm uint16) {
	tempRPM := rpm * 4
	sendFrame(bus, 0x0AA, [8]uint8{
		0xFE,
		0xFE,
		0xFF,
		0x00,
		lo8(tempRPM),
		hi8(tempRPM),
		0xFE,
		0x99,
	})
}
func setFuelLevel(bus *can.Bus, level uint16) {
	sensor := level * 160
	sendFrame(bus, 0x349, [8]uint8{
		lo8(sensor),
		hi8(sensor),
		lo8(sensor),
		hi8(sensor),
		0x00,
	})
}

func sendAirbagSeatbeltCounter(bus *can.Bus) {
	sendFrame(bus, 0x0D7, [8]uint8{
		seatbeltCounter,
		0xFF,
	})
	seatbeltCounter++
}

func sendSeatbletLight(bus *can.Bus) {
	sendFrame(bus, 0x581, [8]uint8{
		0x40,
		0x4d,
		0x00,
		0x28,
		0xFF,
		0xFF,
		0xFF,
		0xFF,
	})
	seatbeltCounter++
}

func sendLigtsOn(bus *can.Bus) {
	sendFrame(bus, 0x21A,
		[8]uint8{
			0x07,
			0x00,
		})
}

func sendHandbrake(handbrakeState bool, bus *can.Bus) {
	handbrake := uint8(0xFE)
	if handbrakeState {
		handbrake = uint8(0xFE)
	} else {
		handbrake = uint8(0xFD)
	}

	sendFrame(bus, 0x34F, [8]uint8{
		handbrake,
		0xFF,
	})
}

func sendAbsBrakeCounter1(bus *can.Bus) {
	sendFrame(bus, 0x0C0, [8]uint8{
		absBrakeCounter1,
		0xFF,
	})
	absBrakeCounter1++
	if absBrakeCounter1 == 0x00 {
		absBrakeCounter1 = 0xF0
	}
}

func sendAbs(bus *can.Bus) {
	absFrame2 = ((((absFrame2 >> 4) + 3) << 4) & 0xF0) | 0x03
	sendFrame(bus, 0x19E, [8]uint8{
		0x00, 0x00, absFrame2, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func sendEngineTemp(bus *can.Bus) {
	engineTemp2++
	sendFrame(bus, 0x1D0, [8]uint8{
		40 + 48, 0xFF, engineTemp2, 0xCD, 0x5D, 0x37, 0xCD, 0xA8,
	})
}

func sendServiceHour(bus *can.Bus) {
	sendFrame(bus, 0x394, [8]uint8{
		0x48, 0x0F, 0x00, 0x0B, 0x00, 0x88, 0x58, 0x01,
	})
}

func sendHazzardLights(bus *can.Bus) {
	sendFrame(bus,
		0x1F6,
		[8]uint8{
			0xB1, 0xF2,
		})
	sendFrame(bus,
		0x1F6,
		[8]uint8{
			0xB1, 0xF1,
		})
}

func sendTime(bus *can.Bus) {
	sendFrame(bus,
		0x39E,
		[8]uint8{
			uint8(time.Now().Hour()),
			uint8(time.Now().Minute()),
			uint8(time.Now().Second()),
			uint8(time.Now().Day()),
			uint8(time.Now().Month()<<4) | 0x0F,
			uint8(time.Now().Year()),
			uint8(time.Now().Year() >> 8),
			0xF2,
		})
}

func sendServiceDistance(bus *can.Bus) {
	sendFrame(bus, 0x394,
		[8]uint8{
			0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		})

}

func send100ms(bus *can.Bus) {
	for {
		ignitionOn(bus)
		ignitionStatus(bus)
		sendRPM(bus, uint16(time.Now().Minute()*100))
		sendSpeed(bus, uint16(int16(time.Now().Hour()*10)-offsetMap[time.Now().Hour()*10]))

		time.Sleep(100 * time.Millisecond)
	}
}

func send200ms(bus *can.Bus) {
	for {
		sendHandbrake(false, bus)
		sendLigtsOn(bus)

		sendSeatbletLight(bus)
		setFuelLevel(bus, 50)
		sendAirbagSeatbeltCounter(bus)
		sendAbsBrakeCounter1(bus)
		sendAbs(bus)
		sendEngineTemp(bus)
		sendServiceHour(bus)
		sendServiceDistance(bus)
		sendTime(bus)
		time.Sleep(200 * time.Millisecond)
	}
}

func send500ms(bus *can.Bus) {
	for {
		if time.Now().Second() == 0 && time.Now().Minute() == 0 && time.Now().Hour() > 7 && time.Now().Hour() < 18 {
			go sendHazzardLights(bus)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func clock(bus *can.Bus) {
	go bus.ConnectAndPublish()
	go send100ms(bus)
	go send200ms(bus)
	go send500ms(bus)
}

func debug(bus *can.Bus, val int) {
	go bus.ConnectAndPublish()
	go send200ms(bus)
	go send500ms(bus)
	for {
		log.Println("sending", uint16(val))
		ignitionOn(bus)
		ignitionStatus(bus)
		sendSpeed(bus, uint16(int16(val)-offsetMap[val]))
		time.Sleep(100 * time.Millisecond)
	}
}

var offsetMap = make(map[int]int16)

func readOffset() {
	offsetsBytes, err := ioutil.ReadFile("/offset.json")
	if err != nil {
		fmt.Println("using default", err)
		offsetMap[10] = 2
		offsetMap[10] = 1
		offsetMap[30] = 2
		offsetMap[40] = 2
		offsetMap[50] = 1
		offsetMap[110] = 1
		offsetMap[120] = 1
		offsetMap[130] = 2
		offsetMap[140] = 2
		offsetMap[150] = 3
		offsetMap[160] = 3
		offsetMap[170] = 3
		offsetMap[180] = 4
		offsetMap[190] = 4
		offsetMap[200] = 5
		offsetMap[210] = 6
		offsetMap[220] = 7
		offsetMap[230] = 7
	}
	err = json.Unmarshal(offsetsBytes, &offsetMap)
	if err == nil {
		fmt.Println("using offset josn")
	}
	if err != nil {
		fmt.Println("using default because", err)
		offsetMap[10] = 2
		offsetMap[10] = 1
		offsetMap[30] = 2
		offsetMap[40] = 2
		offsetMap[50] = 1
		offsetMap[110] = 1
		offsetMap[120] = 1
		offsetMap[130] = 2
		offsetMap[140] = 2
		offsetMap[150] = 3
		offsetMap[160] = 3
		offsetMap[170] = 3
		offsetMap[180] = 4
		offsetMap[190] = 4
		offsetMap[200] = 5
		offsetMap[210] = 6
		offsetMap[220] = 7
		offsetMap[230] = 7
	}
}

func main() {
	readOffset()
	bus, _ := can.NewBusForInterfaceWithName("can0")
	if len(os.Args) <= 1 {
		clock(bus)
	} else {
		log.Println("debug mode")
		val, _ := strconv.Atoi(os.Args[1])
		go debug(bus, val)
	}
	for {
		select {}
	}
}
