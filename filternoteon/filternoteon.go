package filternoteon

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type FilterNoteOn struct {
	channel    filter.FilterChannel
	channelAny bool

	noteAny bool
	note    uint8

	velocityAny bool
	velocity    uint8
}

type FilterNoteOnConfig struct {
	Note     string
	Velocity string
}

const (
	highNibble = 0x90
)

func New(channel filter.FilterChannel, config json.RawMessage) (*FilterNoteOn, error) {
	var f FilterNoteOn
	var conf FilterNoteOnConfig

	f.channel = channel
	err := json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse filter settings :" + err.Error())
	}

	if conf.Note == "*" {
		f.noteAny = true
	} else {
		f.noteAny = false
		value, err := strconv.ParseUint(conf.Note, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid note value: %s", conf.Note)
		}
		f.note = uint8(value)
	}

	if conf.Velocity == "*" {
		f.velocityAny = true
	} else {
		f.velocityAny = false
		value, err := strconv.ParseUint(conf.Velocity, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid note velocity: %s", conf.Velocity)
		}
		f.velocity = uint8(value)
	}

	return &f, nil
}

func (f *FilterNoteOn) String() string {
	var note string
	var velocity string

	if f.noteAny == true {
		note = "*"
	} else {
		note = fmt.Sprintf("%d", f.note)
	}
	if f.velocityAny == true {
		velocity = "*"
	} else {
		velocity = fmt.Sprintf("%d", f.velocity)
	}

	return "Note On on note '" + note + "' with velocity '" + velocity + "'"
}

func (f *FilterNoteOn) QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool {
	if (msgType == filter.FilterMsgTypeNoteOn) && ((f.channel == filter.FilterChannelAny) || (f.channel == channel)) {
		return true
	}

	return false
}

func (f *FilterNoteOn) Match(packet coremidi.Packet) (match filterinterface.FilterMatchResult, value uint16) {
	if len(packet.Data) != 3 || packet.Data[0]&0xF0 != highNibble {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//Note?
	if ((f.noteAny == true) || (packet.Data[1] == f.note)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	//Velocity?
	if ((f.velocityAny == true) || (packet.Data[2] == f.velocity)) == false {
		return filterinterface.FilterMatchResult_NoMatch, 0
	}

	return filterinterface.FilterMatchResult_Match, uint16(packet.Data[2])
}
