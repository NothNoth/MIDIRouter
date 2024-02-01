package genpitchwheel

import (
	"MIDIRouter/filter"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/youpy/go-coremidi"
)

type GenPitchWheel struct {
	channel filter.FilterChannel

	pitchReuse   bool
	pitchReplace bool
	pitch        uint16
}

type FilterPitchWheelConfig struct {
	Pitch string
}

func New(channel filter.FilterChannel, settings json.RawMessage) (*GenPitchWheel, error) {
	var g GenPitchWheel
	var conf FilterPitchWheelConfig

	g.channel = channel

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	g.pitchReuse = false
	g.pitchReplace = false
	if conf.Pitch == "*" {
		g.pitchReuse = true
	} else if conf.Pitch == "$" {
		g.pitchReplace = true
	} else {
		value, err := strconv.ParseUint(conf.Pitch, 10, 16)
		if err != nil {
			return nil, err
		}
		if value > 16383 {
			return nil, fmt.Errorf("Invalid pitch value: %s", conf.Pitch)
		}
		g.pitch = uint16(value)
	}

	return &g, nil
}

func (g *GenPitchWheel) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var statusByte byte
	var pitchLSB byte
	var pitchMSB byte

	filteredMsgType := (packet.Data[0] >> 4)
	filteredChannel := (packet.Data[0] & 0x0F)

	if g.channel == filter.FilterChannelAny {
		statusByte = byte(filter.FilterMsgTypePitchWheel<<4 | filteredChannel)
	} else {
		statusByte = byte(filter.FilterMsgTypePitchWheel<<4 | g.channel)
	}

	//If re-using some values, make sure filtered type is fine
	if (g.pitchReuse == true) && (filteredMsgType != filter.FilterMsgTypePitchWheel) {
		return packet, errors.New("Cannot generate MIDI message with same Pitch, filtered message is of distinct type")
	}

	if g.pitchReuse == true {
		pitchLSB = packet.Data[1]
		pitchMSB = packet.Data[2]
	} else if g.pitchReplace == true {
		//Encode 16bits value into two 7bits
		pitchLSB = byte(value & 0x7F)
		pitchMSB = byte(value >> 7)
	} else {
		pitchLSB = byte(g.pitch & 0x7F)
		pitchMSB = byte(g.pitch >> 7)
	}

	newPacket := coremidi.NewPacket([]byte{statusByte, pitchLSB, pitchMSB}, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenPitchWheel) String() string {
	str := fmt.Sprintf("PitchBend (channel %s) ", g.channel.String())

	if g.pitchReuse == true {
		str += " / set pitch to original pitch value"
	} else if g.pitchReplace == true {
		str += " / set pitch to transformed value"
	} else {
		str += fmt.Sprintf(" / set pitch to %d", g.pitch)
	}

	return str
}
