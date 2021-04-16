package genprogramchange

import (
	"MIDIRouter/filter"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type GenProgramChange struct {
	name    string
	channel filter.FilterChannel

	programNumberReuse   bool
	programNumberReplace bool
	programNumber        uint8
}

type FilterProgramChangeConfig struct {
	ProgramNumber string
}

func New(name string, channel filter.FilterChannel, settings json.RawMessage) (*GenProgramChange, error) {
	var g GenProgramChange
	var conf FilterProgramChangeConfig
	g.name = name
	g.channel = channel

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	g.programNumberReuse = false
	g.programNumberReplace = false
	if conf.ProgramNumber == "*" {
		g.programNumberReuse = true
	} else if conf.ProgramNumber == "$" {
		g.programNumberReplace = true
	} else {
		value, err := strconv.ParseUint(conf.ProgramNumber, 10, 8)
		if err != nil {
			return nil, err
		}
		if value > 127 {
			return nil, fmt.Errorf("Invalid programNumber value: %s", conf.ProgramNumber)
		}
		g.programNumber = uint8(value)
	}

	return &g, nil
}

func (g *GenProgramChange) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var statusByte byte
	var programNumber byte

	filteredMsgType := (packet.Data[0] >> 4)
	filteredChannel := (packet.Data[0] & 0x0F)

	if g.channel == filter.FilterChannelAny {
		statusByte = byte(filter.FilterMsgTypeProgramChange<<4 | filteredChannel)
	} else {
		statusByte = byte(filter.FilterMsgTypeProgramChange<<4 | g.channel)
	}

	//If re-using some values, make sure filtered type is fine
	if (g.programNumberReuse == true) && (filteredMsgType != filter.FilterMsgTypeProgramChange) {
		return packet, errors.New("Cannot generate MIDI message with same ProgramNumber, filtered message is of distinct type")
	}

	if g.programNumberReuse == true {
		programNumber = packet.Data[1]
	} else if g.programNumberReplace == true {
		programNumber = byte(value & 0xFF)
	} else {
		programNumber = g.programNumber
	}

	newPacket := coremidi.NewPacket([]byte{statusByte, programNumber}, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenProgramChange) String() string {
	str := fmt.Sprintf("ProgramChange (channel %s) ", g.channel.String())

	if g.programNumberReuse == true {
		str += " / set program number to original program value"
	} else if g.programNumberReplace == true {
		str += " / set program number to transformed value"
	} else {
		str += fmt.Sprintf(" / set program number to %d", g.programNumber)
	}

	return str
}
