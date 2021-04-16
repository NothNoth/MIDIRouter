package gennoteon

import (
	"MIDIRouter/filter"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type GenNoteOn struct {
	name    string
	channel filter.FilterChannel

	noteReuse   bool
	noteReplace bool
	note        uint8

	velocityReuse   bool
	velocityReplace bool
	velocity        uint8
}

type FilterNoteOnConfig struct {
	Note     string
	Velocity string
}

func New(name string, channel filter.FilterChannel, settings json.RawMessage) (*GenNoteOn, error) {
	var g GenNoteOn
	var conf FilterNoteOnConfig
	g.name = name
	g.channel = channel

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	g.noteReuse = false
	g.noteReplace = false
	if conf.Note == "*" {
		g.noteReuse = true
	} else if conf.Note == "$" {
		g.noteReplace = true
	} else {
		value, err := strconv.ParseUint(conf.Note, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid note value: %s", conf.Note)
		}
		g.note = uint8(value)
	}

	g.velocityReuse = false
	g.velocityReplace = false
	if conf.Velocity == "*" {
		g.velocityReuse = true
	} else if conf.Velocity == "$" {
		g.velocityReplace = true
	} else {
		value, err := strconv.ParseUint(conf.Velocity, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid note velocity: %s", conf.Velocity)
		}
		g.velocity = uint8(value)
	}

	return &g, nil
}

func (g *GenNoteOn) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var statusByte byte
	var note byte
	var velocity byte

	filteredMsgType := (packet.Data[0] >> 4)
	filteredChannel := (packet.Data[0] & 0x0F)

	if g.channel == filter.FilterChannelAny {
		statusByte = byte(filter.FilterMsgTypeNoteOn<<4 | filteredChannel)
	} else {
		statusByte = byte(filter.FilterMsgTypeNoteOn<<4 | g.channel)
	}

	//If re-using some values, make sure filtered type is fine
	if (g.noteReuse == true) || (g.velocityReuse == true) {
		if (filteredMsgType != filter.FilterMsgTypeNoteOn) && (filteredMsgType != filter.FilterMsgTypeNoteOff) {
			return packet, errors.New("Cannot generate MIDI message with same Note/Velocity, filtered message is of distinct type")
		}
	}

	if g.noteReuse == true {
		note = packet.Data[1]
	} else if g.noteReplace == true {
		note = byte(value & 0xFF)
	} else {
		note = g.note
	}

	if g.velocityReuse == true {
		velocity = packet.Data[2]
	} else if g.velocityReplace == true {
		velocity = byte(value & 0xFF)
	} else {
		velocity = g.velocity
	}

	newPacket := coremidi.NewPacket([]byte{statusByte, note, velocity}, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenNoteOn) String() string {
	str := fmt.Sprintf("NoteOn (channel %s) ", g.channel.String())

	if g.noteReuse == true {
		str += " / set note to original note value"
	} else if g.noteReplace == true {
		str += " / set note to transformed value"
	} else {
		str += fmt.Sprintf(" / set note to %d", g.note)
	}

	if g.velocityReuse == true {
		str += " / set velocity to original velocity value"
	} else if g.velocityReplace == true {
		str += " / set velocity to transformed value"
	} else {
		str += fmt.Sprintf(" / set velocity to %d", g.velocity)
	}

	return str
}
