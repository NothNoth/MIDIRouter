package filteraftertouch

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type FilterAftertouch struct {
	channel    filter.FilterChannel
	channelAny bool

	pressureAny bool
	pressure    uint8
}

type FilterAftertouchConfig struct {
	Pressure string
}

func New(channel filter.FilterChannel, config json.RawMessage) (*FilterAftertouch, error) {
	var f FilterAftertouch
	var conf FilterAftertouchConfig

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
			return nil, fmt.Errorf("Invalid note pressure: %s", conf.Pressure)
		}
		f.pressure = uint8(value)
	}

	return &f, nil
}

func (f *FilterAftertouch) String() string {
	var pressure string

	if f.pressureAny == true {
		pressure = "*"
	} else {
		pressure = fmt.Sprintf("%d", f.pressure)
	}

	return "Aftertouch with pressure '" + pressure + "'"
}

func (f *FilterAftertouch) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypeAftertouch) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterAftertouch) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 2 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//pressure ?
	pressureMatch := ((f.pressureAny == true) || (packet.Data[1] == f.pressure))
	if pressureMatch == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, uint16(packet.Data[1])
}
