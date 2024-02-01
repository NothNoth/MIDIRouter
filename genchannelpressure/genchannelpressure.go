package genchannelpressure

import (
	"MIDIRouter/filter"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type GenChannelPressure struct {
	channel filter.FilterChannel

	pressureReuse   bool
	pressureReplace bool
	pressure        uint8
}

type FilterChannelPressureConfig struct {
	Pressure string
}

func New(channel filter.FilterChannel, settings json.RawMessage) (*GenChannelPressure, error) {
	var g GenChannelPressure
	var conf FilterChannelPressureConfig

	g.channel = channel

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	g.pressureReuse = false
	g.pressureReplace = false
	if conf.Pressure == "*" {
		g.pressureReuse = true
	} else if conf.Pressure == "$" {
		g.pressureReplace = true
	} else {
		value, err := strconv.ParseUint(conf.Pressure, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid pressure value: %s", conf.Pressure)
		}
		g.pressure = uint8(value)
	}

	return &g, nil
}

func (g *GenChannelPressure) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var statusByte byte
	var pressure byte

	filteredMsgType := (packet.Data[0] >> 4)
	filteredChannel := (packet.Data[0] & 0x0F)

	if g.channel == filter.FilterChannelAny {
		statusByte = byte(filter.FilterMsgTypeChannelPressure<<4 | filteredChannel)
	} else {
		statusByte = byte(filter.FilterMsgTypeChannelPressure<<4 | g.channel)
	}

	//If re-using some values, make sure filtered type is fine
	if g.pressureReuse == true && (filteredMsgType != filter.FilterMsgTypeChannelPressure) && (filteredMsgType != filter.FilterMsgTypeAftertouch) {
		return packet, errors.New("Cannot generate MIDI message with same Pressure, filtered message is of distinct type")
	}

	if g.pressureReuse == true {
		pressure = packet.Data[1]
	} else if g.pressureReplace == true {
		pressure = byte(value & 0xFF)
	} else {
		pressure = g.pressure
	}

	newPacket := coremidi.NewPacket([]byte{statusByte, pressure}, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenChannelPressure) String() string {
	str := fmt.Sprintf("ChannelPressure (channel %s) ", g.channel.String())
	if g.pressureReuse == true {
		str += " - set pressure to original pressure value"
	} else if g.pressureReplace == true {
		str += " - set pressure to transformed value"
	} else {
		str += fmt.Sprintf(" / set pressure to %d", g.pressure)
	}

	return str
}
