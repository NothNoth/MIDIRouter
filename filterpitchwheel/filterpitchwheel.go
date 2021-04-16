package filterpitchwheel

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type FilterPitchWheel struct {
	name string

	channel    filter.FilterChannel
	channelAny bool

	pitchAny bool
	pitch    uint16
}

type FilterPitchWheelConfig struct {
	Pitch string
}

func New(name string, channel filter.FilterChannel, config json.RawMessage) (*FilterPitchWheel, error) {
	var f FilterPitchWheel
	var conf FilterPitchWheelConfig

	f.name = name
	f.channel = channel
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse filter settings :" + err.Error())
	}

	if conf.Pitch == "*" {
		f.pitchAny = true
	} else {
		f.pitchAny = false

		value, err := strconv.ParseUint(conf.Pitch, 10, 16)
		if err != nil {
			return nil, err
		}
		if value > 0xFFFF {
			return nil, errors.New("Pitch value too large")
		}
		f.pitch = uint16(value)
	}

	return &f, nil
}

func (f *FilterPitchWheel) String() string {
	var pitch string

	if f.pitchAny == true {
		pitch = "*"
	} else {
		pitch = fmt.Sprintf("%d", f.pitch)
	}

	return "Pitch Wheel change with value '" + pitch + "'"
}

func (f *FilterPitchWheel) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypePitchWheel) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterPitchWheel) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 3 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}
	low := uint16(packet.Data[1])
	high := uint16(packet.Data[2])
	value = (high << 7) | low

	if f.pitchAny == true {
		return filterinterface.FilterMatchResult_Match, value
	}

	//Pitch?
	if (value == f.pitch) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}
	return filterinterface.FilterMatchResult_Match, value
}
