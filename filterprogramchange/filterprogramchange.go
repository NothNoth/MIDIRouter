package filterprogramchange

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type FilterProgramChange struct {
	channel    filter.FilterChannel
	channelAny bool

	programNumberAny bool
	programNumber    uint8
}

type FilterProgramChangeConfig struct {
	ProgramNumber string
}

func New(channel filter.FilterChannel, config json.RawMessage) (*FilterProgramChange, error) {
	var f FilterProgramChange
	var conf FilterProgramChangeConfig

	f.channel = channel
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse filter settings :" + err.Error())
	}

	if conf.ProgramNumber == "*" {
		f.programNumberAny = true
	} else {
		f.programNumberAny = false
		value, err := strconv.ParseUint(conf.ProgramNumber, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid program number %s", conf.ProgramNumber)
		}
		f.programNumber = uint8(value)
	}

	return &f, nil
}

func (f *FilterProgramChange) String() string {
	var programNumber string

	if f.programNumberAny == true {
		programNumber = "*"
	} else {
		programNumber = fmt.Sprintf("%02x", f.programNumber)
	}

	return "Program Change '" + programNumber + "'"
}

func (f *FilterProgramChange) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypeProgramChange) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterProgramChange) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 2 {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//ProgramNumber?
	if ((f.programNumberAny == true) || (packet.Data[1] == f.programNumber)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, uint16(packet.Data[1])
}
