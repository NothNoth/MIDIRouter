package rule

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"MIDIRouter/generatorinterface"
	"errors"
	"fmt"
	"time"

	"github.com/youpy/go-coremidi"
)

type RuleMatchResult int

const (
	RuleMatchResultNoMatch       = iota
	RuleMatchResultMatchInject   = iota
	RuleMatchResultMatchNoInject = iota
)

type TransformMode uint8

const (
	TransformModeNone       = iota
	TransformModeLinear     = iota
	TransformModeLinearDrop = iota
	/*
		Transform7bitsTo100   = iota // Convert a 7bits [0-127] value to a displayed [0-100] value
		Transform14bitsTo100  = iota // Convert a 14bits [0-16383] value to a displayed [0-100] value
		Transform7bitsTo1000  = iota // Convert a 7bits [0-127] value to a displayed [0-1000] value
		Transform14bitsTo1000 = iota // Convert a 14bits [0-16383] value to a displayed [0-1000] value
		Transform7bitsTo127   = iota // Convert a 7bits [0-127] value to a [0-127] value
		Transform14bitsTo127  = iota // Convert a 14bits [0-16383] value to a [0-127] value*/
)

type Transform struct {
	mode    TransformMode
	fromMin uint32
	fromMax uint32
	toMin   uint32
	toMax   uint32
}

type Rule struct {
	name                  string
	filter                filterinterface.FilterInterface
	transform             Transform
	dropDuplicates        bool
	dropDuplicatesTimeout time.Duration

	generator generatorinterface.GeneratorInterface

	lastValue   uint16
	lastValueTs time.Time
}

func New(ruleName string) (*Rule, error) {
	var r Rule

	r.name = ruleName
	r.lastValue = 0xFFFF
	r.transform.mode = TransformModeNone
	return &r, nil
}

func (r *Rule) SetTransform(mode TransformMode, fromMin uint32, fromMax uint32, toMin uint32, toMax uint32) {
	r.transform = Transform{mode: mode, fromMin: fromMin, fromMax: fromMax, toMin: toMin, toMax: toMax}
}

func (r *Rule) SetFilter(f filterinterface.FilterInterface) error {
	if r.filter != nil {
		return errors.New("Filter already set")
	}
	r.filter = f
	return nil
}

func (r *Rule) EnableDropDuplicates(enable bool, timeout time.Duration) {
	r.dropDuplicates = enable
	r.dropDuplicatesTimeout = timeout
}

func (r *Rule) SetGenerator(g generatorinterface.GeneratorInterface) error {
	if r.generator != nil {
		return errors.New("Generator already set")
	}
	r.generator = g
	return nil
}

func (r *Rule) Match(packet coremidi.Packet) (RuleMatchResult, coremidi.Packet) {
	msgType := filter.FilterMsgType((packet.Data[0] & 0xF0) >> 4)
	channel := filter.FilterChannel(packet.Data[0] & 0x0F)
	if r.filter.QuickMatch(msgType, channel) == false {
		return RuleMatchResultNoMatch, packet
	}

	result, value := r.filter.Match(packet)
	if result == filterinterface.FilterMatchResult_NoMatch {
		return RuleMatchResultNoMatch, packet
	}

	if result == filterinterface.FilterMatchResult_MatchNoValue {
		fmt.Println("Filter match (no value)")
		return RuleMatchResultMatchNoInject, packet
	}
	if result != filterinterface.FilterMatchResult_Match {
		return RuleMatchResultNoMatch, packet
	}

	fmt.Println("Filter", r.filter, "matched. Extracted value:", value)
	fmt.Println("-> Extracted value:", value)

	//Transform
	switch r.transform.mode {
	case TransformModeLinear:
		//Alright, this is serious mathematics.
		//Transform 'value' which should be in the range [r.transform.fromMin, r.transform.fromMax]
		//										   to a value in the range [r.transform.toMin, r.transform.toMax]
		// A(r.transform.fromMin, r.transform.toMin) / B(r.transform.fromMax, r.transform.toMax)
		a := float64(r.transform.toMax-r.transform.toMin) / float64(r.transform.fromMax-r.transform.fromMin)
		//y = ax + b => b = y - ax
		b := float64(r.transform.toMin) - a*float64(r.transform.fromMin)
		value = uint16(a*float64(value) + float64(b))
	case TransformModeLinearDrop:
		//Input out of bounds
		if (uint32(value) > r.transform.fromMax) || (uint32(value) < r.transform.fromMin) {
			fmt.Println("-> Tranform dropped out of bounds input value")
			return RuleMatchResultNoMatch, packet
		}
		//Alright, this is serious mathematics.
		//Transform 'value' which should be in the range [r.transform.fromMin, r.transform.fromMax]
		//										   to a value in the range [r.transform.toMin, r.transform.toMax]
		// A(r.transform.fromMin, r.transform.toMin) / B(r.transform.fromMax, r.transform.toMax)
		a := float64(r.transform.toMax-r.transform.toMin) / float64(r.transform.fromMax-r.transform.fromMin)
		//y = ax + b => b = y - ax
		b := float64(r.transform.toMin) - a*float64(r.transform.fromMin)
		v := uint16(a*float64(value) + float64(b))
		//Input out of bounds
		if (uint32(v) > r.transform.toMax) || (uint32(v) < r.transform.toMin) {
			fmt.Println("-> Tranform dropped out of bounds output value")
			return RuleMatchResultNoMatch, packet
		}
		value = v
	case TransformModeNone:
	default:
	}
	/*
		switch r.transform {
		case Transform7bitsTo100:
			value = uint16((float32(value) * 100.0) / 0x7F)
		case Transform14bitsTo100:
			value = uint16((float32(value) * 100.0) / 0x3FFF)
		case Transform7bitsTo1000:
			value = uint16((float32(value) * 1000.0) / 0x7F)
		case Transform14bitsTo1000:
			value = uint16((float32(value) * 1000.0) / 0x3FFF)
		case Transform14bitsTo127:
			value = uint16((float32(value) * 127.0) / 0x3FFF)
		case TransformNone:
		case Transform7bitsTo127:
		default:
		}
		if r.transform != TransformNone {
			fmt.Println("-> Converted value: ", value)
		}*/

	//Drop?
	if r.dropDuplicates && (r.lastValue == value) && (time.Since(r.lastValueTs) < r.dropDuplicatesTimeout) {
		fmt.Println("-> Ignored duplicate")
		return RuleMatchResultMatchNoInject, packet
	}
	r.lastValue = value
	r.lastValueTs = time.Now()

	//Generate output
	newPacket, err := r.output(packet, value)
	if err != nil {
		fmt.Println(err)
		return RuleMatchResultMatchInject, packet
	} else {
		return RuleMatchResultMatchInject, newPacket
	}
	return RuleMatchResultMatchInject, packet

}

func (r *Rule) output(packet coremidi.Packet, value uint16) (newPacket coremidi.Packet, err error) {

	newPacket, err = r.generator.Generate(packet, value)
	if err != nil {
		return packet, err
	}

	return newPacket, nil
}

func (r Rule) String() string {
	var str string
	str += "***** Rule '" + r.name + "' *****\n"
	str += "  Match    : " + r.filter.String() + "\n"
	str += "  Transform: " + r.transform.String() + "\n"
	str += "  Output   : " + r.generator.String()

	return str
}

func (t Transform) String() string {

	switch t.mode {
	case TransformModeNone:
		return "None"
	case TransformModeLinear:
		return fmt.Sprintf("Linear from [%d, %d] to [%d, %d]", t.fromMin, t.fromMax, t.toMin, t.toMax)
	case TransformModeLinearDrop:
		return fmt.Sprintf("Linear from [%d, %d] to [%d, %d] (drop out of range values)", t.fromMin, t.fromMax, t.toMin, t.toMax)
	}
	return "?"
}
