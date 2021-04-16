package gencontrolchange

import (
	"MIDIRouter/filter"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type ControlChangeMode uint8

const (
	controlChangeModeStandard ControlChangeMode = iota
	controlChangeModeCCAh     ControlChangeMode = iota
)

type GenControlChange struct {
	name    string
	channel filter.FilterChannel
	mode    ControlChangeMode

	controllerNumberReuse   bool
	controllerNumberReplace bool
	controllerNumber        uint8

	valueReuse   bool
	valueReplace bool
	value        uint8
}

type FilterControlChangeConfig struct {
	Mode             string
	ControllerNumber string
	Value            string
}

func New(name string, channel filter.FilterChannel, settings json.RawMessage) (*GenControlChange, error) {
	var g GenControlChange
	var conf FilterControlChangeConfig
	g.name = name
	g.channel = channel

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	if conf.Mode == "Standard" {
		g.mode = controlChangeModeStandard
	} else if conf.Mode == "CCAh" {
		g.mode = controlChangeModeCCAh
	} else if len(conf.Mode) != 0 {
		return nil, errors.New("Failed to parse generator settings: invalid mode " + conf.Mode)
	}

	g.controllerNumberReuse = false
	g.controllerNumberReplace = false
	if conf.ControllerNumber == "*" {
		g.controllerNumberReuse = true
	} else if conf.ControllerNumber == "$" {
		g.controllerNumberReplace = true
	} else {
		value, err := strconv.ParseUint(conf.ControllerNumber, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid controllerNumber value: %s", conf.ControllerNumber)
		}
		g.controllerNumber = uint8(value)
	}

	g.valueReuse = false
	g.valueReplace = false
	if conf.Value == "*" {
		g.valueReuse = true
	} else if conf.Value == "$" {
		g.valueReplace = true
	} else {
		value, err := strconv.ParseUint(conf.Value, 10, 8)
		if err != nil {
			return nil, err
		}

		if (g.mode == controlChangeModeStandard) && (value > 127) {
			return nil, fmt.Errorf("Invalid value: %s", conf.Value)
		} else if (g.mode == controlChangeModeCCAh) && (value > 16383) {
			return nil, fmt.Errorf("Invalid value: %s", conf.Value)
		}
		g.value = uint8(value)
	}

	return &g, nil
}

func (g *GenControlChange) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	if g.mode == controlChangeModeStandard {
		return g.generateStandard(packet, value)
	} else if g.mode == controlChangeModeCCAh {
		return g.generateCCAh(packet, value)
	}

	return packet, errors.New("Invalid generate mode")
}

func (g *GenControlChange) generateStandard(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var statusByte byte
	var controllerNumber byte
	var newValue byte

	filteredMsgType := (packet.Data[0] >> 4)
	filteredChannel := (packet.Data[0] & 0x0F)

	if g.channel == filter.FilterChannelAny {
		statusByte = byte(filter.FilterMsgTypeControlChange<<4 | filteredChannel)
	} else {
		statusByte = byte(filter.FilterMsgTypeControlChange<<4 | g.channel)
	}

	//If re-using some values, make sure filtered type is fine
	if ((g.controllerNumberReuse == true) || (g.valueReuse == true)) && (filteredMsgType != filter.FilterMsgTypeControlChange) {
		return packet, errors.New("Cannot generate MIDI message with same ControllerNumber/Value, filtered message is of distinct type")
	}

	if g.controllerNumberReuse == true {
		controllerNumber = packet.Data[1]
	} else if g.controllerNumberReplace == true {
		controllerNumber = byte(value & 0xFF)
	} else {
		controllerNumber = g.controllerNumber
	}

	if g.valueReuse == true {
		newValue = packet.Data[2]
	} else if g.valueReplace == true {
		newValue = byte(value & 0xFF)
	} else {
		newValue = g.value
	}

	newPacket := coremidi.NewPacket([]byte{statusByte, controllerNumber, newValue}, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenControlChange) generateCCAh(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {

	// We need to return 2 packets
	// On reuse value, we only have the last message value
	return packet, errors.New("Not implemented")
}

func (g *GenControlChange) String() string {
	str := fmt.Sprintf("ControlChange (channel %s) ", g.channel.String())

	switch g.mode {
	case controlChangeModeStandard:
		str += " / Mode: standard"
	case controlChangeModeCCAh:
		str += " / Mode: CCAh (14bits)"
	}

	if g.controllerNumberReuse == true {
		str += " / set control number to original value"
	} else if g.controllerNumberReplace {
		str += " / set control number to transformed value"
	} else {
		str += fmt.Sprintf(" / set control number to %d", g.controllerNumber)
	}

	if g.valueReuse == true {
		str += " / set value to original value"
	} else if g.valueReplace == true {
		str += " / set value to transformed value"
	} else {
		str += fmt.Sprintf(" / set value to %d", g.value)
	}

	return str
}
