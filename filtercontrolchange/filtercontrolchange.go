package filtercontrolchange

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
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

type FilterControlChange struct {
	mode       ControlChangeMode
	channel    filter.FilterChannel
	channelAny bool

	controllerNumberAny bool
	controllerNumber    uint8

	valueAny bool
	value    uint16

	ccahFlag  bool
	ccahValue uint8
}

type FilterControlChangeConfig struct {
	ControllerNumber string
	Value            string
	Mode             string
}

func New(channel filter.FilterChannel, config json.RawMessage) (*FilterControlChange, error) {
	var f FilterControlChange
	var conf FilterControlChangeConfig

	f.mode = controlChangeModeStandard
	f.channel = channel
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse filter settings :" + err.Error())
	}

	if conf.Mode == "Standard" {
		f.mode = controlChangeModeStandard
	} else if conf.Mode == "CCAh" {
		f.mode = controlChangeModeCCAh
	} else if len(conf.Mode) != 0 {
		return nil, errors.New("Failed to parse filter settings: invalid mode " + conf.Mode)
	}

	if conf.ControllerNumber == "*" {
		f.controllerNumberAny = true
	} else {
		f.controllerNumberAny = false
		value, err := strconv.ParseUint(conf.ControllerNumber, 10, 8)
		if err != nil {
			return nil, err
		}
		if (f.mode == controlChangeModeStandard) && (value > 127) {
			return nil, fmt.Errorf("Invalid controller number value: %s", conf.ControllerNumber)
		} else if (f.mode == controlChangeModeCCAh) && (value > 31) {
			return nil, fmt.Errorf("Invalid controller number value: %s", conf.ControllerNumber)
		}
		f.controllerNumber = uint8(value)
	}

	if conf.Value == "*" {
		f.valueAny = true
	} else {
		f.valueAny = false
		value, err := strconv.ParseUint(conf.Value, 10, 16)
		if err != nil {
			return nil, err
		}
		if (f.mode == controlChangeModeStandard) && (value > 127) {
			return nil, fmt.Errorf("Invalid value: %s", conf.Value)
		} else if (f.mode == controlChangeModeCCAh) && (value > 16383) {
			return nil, fmt.Errorf("Invalid value: %s", conf.Value)
		}
		f.value = uint16(value)
	}

	f.ccahFlag = false
	return &f, nil
}

func (f *FilterControlChange) String() string {
	var controllerNumber string
	var value string

	if f.controllerNumberAny == true {
		controllerNumber = "*"
	} else {
		controllerNumber = fmt.Sprintf("%d", f.controllerNumber)
	}
	if f.valueAny == true {
		value = "*"
	} else {
		value = fmt.Sprintf("%d", f.value)
	}

	return "Control Change on controller '" + controllerNumber + "' with value '" + value + "' (mode: " + modeToString(f.mode) + ")"
}

func modeToString(mode ControlChangeMode) string {
	switch mode {
	case controlChangeModeStandard:
		return "Standard"
	case controlChangeModeCCAh:
		return "CCAh"
	}
	return "Unknown"
}

func (f *FilterControlChange) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypeControlChange) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterControlChange) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 3 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	if f.mode == controlChangeModeStandard {
		return f.matchStandard(packet)
	} else if f.mode == controlChangeModeCCAh {
		return f.matchCCAh(packet)
	}

	//Should not be here.
	return filterinterface.FilterMatchResult_NoMatch, 0
}

func (f *FilterControlChange) matchStandard(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 3 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}
	//ControllerNumber?
	if ((f.controllerNumberAny == true) || (packet.Data[1] == f.controllerNumber)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//Value?
	if ((f.valueAny == true) || (packet.Data[2] == uint8(f.value))) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, uint16(packet.Data[2])
}

func (f *FilterControlChange) matchCCAh(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 3 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//Read MSB bits : just match on CC Number and wait for second message and full value
	if f.ccahFlag == false {
		//ControllerNumber?
		if ((f.controllerNumberAny == true) || (packet.Data[1] == f.controllerNumber)) == false {
			return filterinterface.FilterMatchResult_NoMatch, 0
		}

		//Match, wait for second message
		f.ccahFlag = true
		f.ccahValue = uint8(packet.Data[2])
		return filterinterface.FilterMatchResult_MatchNoValue, 0
	}

	//Read LSB bits
	f.ccahFlag = false

	//ControllerNumber?
	if ((f.controllerNumberAny == true) || ((packet.Data[1] - 0x20) == f.controllerNumber)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}
	value = uint16(f.ccahValue)<<7 | uint16(packet.Data[2])

	//Value?
	if ((f.valueAny == true) || (value == f.value)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, value
}
