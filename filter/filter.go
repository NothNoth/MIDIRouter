package filter

type FilterMsgType uint8

const (
	FilterMsgTypeUnknown         = 0x0
	FilterMsgTypeNoteOn          = 0x8
	FilterMsgTypeNoteOff         = 0x9
	FilterMsgTypeAftertouch      = 0xA
	FilterMsgTypeControlChange   = 0xB
	FilterMsgTypeProgramChange   = 0xC
	FilterMsgTypeChannelPressure = 0xD
	FilterMsgTypePitchWheel      = 0xE
	FilterMsgTypeSysEx           = 0xF0
	FilterMsgTypeAny             = 0xFF
)

type FilterChannel uint8

const (
	FilterChannel1   = 0x00
	FilterChannel2   = 0x01
	FilterChannel3   = 0x02
	FilterChannel4   = 0x03
	FilterChannel5   = 0x04
	FilterChannel6   = 0x05
	FilterChannel7   = 0x06
	FilterChannel8   = 0x07
	FilterChannel9   = 0x08
	FilterChannel10  = 0x09
	FilterChannel11  = 0x0A
	FilterChannel12  = 0x0B
	FilterChannel13  = 0x0C
	FilterChannel14  = 0x0D
	FilterChannel15  = 0x0E
	FilterChannel16  = 0x0F
	FilterChannelAny = 0xFF
)

func (s FilterMsgType) String() string {
	switch s {
	case FilterMsgTypeNoteOn:
		return "Note On"
	case FilterMsgTypeNoteOff:
		return "Note Off"
	case FilterMsgTypeAftertouch:
		return "Aftertouch"
	case FilterMsgTypeControlChange:
		return "Control Change"
	case FilterMsgTypeProgramChange:
		return "Program Change"
	case FilterMsgTypeChannelPressure:
		return "Channel Pressure"
	case FilterMsgTypePitchWheel:
		return "Pitch Wheel"
	case FilterMsgTypeSysEx:
		return "SysEx"
	case FilterMsgTypeAny:
		return "*"
	default:
		return "Unknown"
	}
}

func (c FilterChannel) String() string {
	switch c {
	case FilterChannel1:
		return "1"
	case FilterChannel2:
		return "2"
	case FilterChannel3:
		return "3"
	case FilterChannel4:
		return "4"
	case FilterChannel5:
		return "5"
	case FilterChannel6:
		return "6"
	case FilterChannel7:
		return "7"
	case FilterChannel8:
		return "8"
	case FilterChannel9:
		return "9"
	case FilterChannel10:
		return "10"
	case FilterChannel11:
		return "11"
	case FilterChannel12:
		return "12"
	case FilterChannel13:
		return "13"
	case FilterChannel14:
		return "14"
	case FilterChannel15:
		return "15"
	case FilterChannel16:
		return "16"
	case FilterChannelAny:
		return "*"
	default:
		return "Unknown"
	}
}
