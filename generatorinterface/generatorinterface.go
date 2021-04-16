package generatorinterface

import "github.com/youpy/go-coremidi"

type GeneratorInterface interface {
	Generate(packet coremidi.Packet, value uint16) (generate coremidi.Packet, err error)
	String() string
}
