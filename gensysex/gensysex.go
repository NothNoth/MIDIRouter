package gensysex

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/youpy/go-coremidi"
)

type Mode int

const (
	Mode7Bits         = iota
	Mode14Bits        = iota
	ModeEnsoniq14To32 = iota
)

type GenSysEx struct {
	name   string
	mode   Mode
	prefix []byte
	suffix []byte
}

type FilterSysExConfig struct {
	Prefix string
	Suffix string
	Mode   string
}

func New(name string, settings json.RawMessage) (*GenSysEx, error) {
	var g GenSysEx
	var conf FilterSysExConfig
	g.name = name

	err := json.Unmarshal([]byte(settings), &conf)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}

	switch conf.Mode {
	case "7bits":
		g.mode = Mode7Bits
	case "14bits":
		g.mode = Mode14Bits
	case "Ensoniq14To32":
		g.mode = ModeEnsoniq14To32
	default:
		return nil, errors.New("invalid convert mode: " + conf.Mode)
	}

	data, err := hex.DecodeString(conf.Prefix)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}
	g.prefix = append(g.prefix, data...)

	if (len(g.prefix) == 0) || (g.prefix[0] != 0xF0) {
		return nil, errors.New("Invalid SysEx prefix, must start with F0")
	}

	data, err = hex.DecodeString(conf.Suffix)
	if err != nil {
		return nil, errors.New("Failed to parse generator settings :" + err.Error())
	}
	g.suffix = append(g.suffix, data...)
	if (len(g.suffix) == 0) || (g.suffix[0] != 0xF7) {
		return nil, errors.New("Invalid SysEx prefix, must end with F7")
	}

	return &g, nil
}

func (g *GenSysEx) Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error) {
	var data []byte

	data = append(data, g.prefix...)
	switch g.mode {
	case Mode7Bits:
		data = append(data, byte(value))
	case Mode14Bits:
		pitchLSB := byte(value & 0x7F)
		pitchMSB := byte(value >> 7)
		data = append(data, pitchLSB)
		data = append(data, pitchMSB)
	case ModeEnsoniq14To32:
		a := byte((value >> 12) & 0xF)
		b := byte((value >> 8) & 0xF)
		c := byte((value >> 4) & 0xF)
		d := byte((value >> 0) & 0xF)
		data = append(data, a)
		data = append(data, b)
		data = append(data, c)
		data = append(data, d)
	}

	data = append(data, g.suffix...)

	newPacket := coremidi.NewPacket(data, packet.TimeStamp)

	return newPacket, nil
}

func (g *GenSysEx) String() string {
	str := "Sysex"

	str += " / " + hex.EncodeToString(g.prefix)
	switch g.mode {
	case Mode7Bits:
		str += "[7bits value]"
	case Mode14Bits:
		str += "[14bits value]"
	case ModeEnsoniq14To32:
		str += "[Ensoniq 14 to 32 bits]"
	}
	str += hex.EncodeToString(g.suffix)

	return str
}
