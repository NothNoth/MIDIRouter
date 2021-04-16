package filterinterface

import (
	"MIDIRouter/filter"

	"github.com/youpy/go-coremidi"
)

type FilterMatchResult int

const (
	FilterMatchResult_Match        = iota
	FilterMatchResult_MatchNoValue = iota
	FilterMatchResult_NoMatch      = iota
)

type FilterInterface interface {
	QuickMatch(msgType filter.FilterMsgType, channel filter.FilterChannel) bool
	Match(packet coremidi.Packet) (match FilterMatchResult, value uint16)
	String() string
}
