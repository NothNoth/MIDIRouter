package filterchannelpressure

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type FilterChannelPressure struct {
	channel    filter.FilterChannel
	channelAny bool

	pressureAny bool
	pressure    uint8
}

type FilterChannelPressureConfig struct {
	Pressure string
}

const (
	highNibble = 0xD0
)

func New(channel filter.FilterChannel, config json.RawMessage) (*FilterChannelPressure, error) {
	var f FilterChannelPressure
	var conf FilterChannelPressureConfig

	f.channel = channel
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse filter settings :" + err.Error())
	}

	if conf.Pressure == "*" {
		f.pressureAny = true
	} else {
		f.pressureAny = false
		value, err := strconv.ParseUint(conf.Pressure, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid pressure value: %s", conf.Pressure)
		}
		f.pressure = uint8(value)
	}

	return &f, nil
}

func (f *FilterChannelPressure) String() string {
	var pressure string

	if f.pressureAny == true {
		pressure = "*"
	} else {
		pressure = fmt.Sprintf("%02x", f.pressure)
	}

	return "Channel pressure '" + pressure + "'"
}

func (f *FilterChannelPressure) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypeChannelPressure) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterChannelPressure) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 2 || packet.Data[0]&0xF0 != highNibble {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//Pressure?
	if ((f.pressureAny == true) || (packet.Data[1] == f.pressure)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, uint16(packet.Data[1])
}
