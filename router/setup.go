package router

import (
	"errors"
	"fmt"
	"strings"

	"github.com/youpy/go-coremidi"
)

func (relay *MIDIRouter) setupSource() error {
	split := strings.Split(relay.sourceDevice, "/")
	if len(split) != 2 {
		return errors.New("invalid destination format, expected: <manufacturer>/<port> got " + relay.sourceDevice)
	}
	manufacturer := split[0]
	port := split[1]

	//Setup MIDI source
	valid, err := isValidSourceDevice(manufacturer, port)
	if err != nil {
		return err
	}
	if valid == false {
		return errors.New("Invalid MIDI source: " + relay.sourceDevice)
	}
	relay.srcPort, err = coremidi.NewInputPort(relay.midiClient, relay.sourceDevice+" input port",
		func(source coremidi.Source, packet coremidi.Packet) {
			relay.onPacket(source, packet)
		})
	if err != nil {
		return err
	}

	sources, err := coremidi.AllSources()
	if err != nil {
		panic(err)
	}
	found := false
	for _, source := range sources {
		if source.Manufacturer() == manufacturer && source.Name() == port {
			relay.srcPort.Connect(source)
			fmt.Println("Source device: ", source.Entity().Device().Name(), "(", source.Manufacturer(), ")")
			found = true
			break
		}
	}

	if found == false {
		return errors.New("MIDI source not found.")
	}
	return nil
}

func (relay *MIDIRouter) setupDestination() error {
	split := strings.Split(relay.destinationDevice, "/")
	if len(split) != 2 {
		return errors.New("invalid destination format, expected: <manufacturer>/<port> got " + relay.destinationDevice)
	}
	manufacturer := split[0]
	port := split[1]

	//Setup MIDI destination
	valid, err := isValidDestinationDevice(manufacturer, port)
	if err != nil {
		return err
	}
	if valid == false {
		return errors.New("Invalid MIDI destination: " + relay.destinationDevice)
	}
	relay.destPort, err = coremidi.NewOutputPort(relay.midiClient, relay.destinationDevice+" output port")
	if err != nil {
		return err
	}
	destinations, err := coremidi.AllDestinations()
	if err != nil {
		return err
	}
	found := false
	for _, destination := range destinations {
		if destination.Manufacturer() == manufacturer && destination.Name() == port {
			relay.destination = destination
			fmt.Println("Destination device: ", destination.Name(), "(", destination.Manufacturer(), ")")
			found = true
			break
		}
	}
	if found == false {
		return errors.New("MIDI destination not found")
	}
	return nil
}

func isValidSourceDevice(manufactuer string, name string) (bool, error) {
	sources, err := coremidi.AllSources()
	if err != nil {
		return false, err
	}

	for _, s := range sources {
		if s.Manufacturer() == manufactuer && s.Name() == name {
			return true, nil
		}
	}
	return false, nil
}

func isValidDestinationDevice(manufactuer string, name string) (bool, error) {
	destinations, err := coremidi.AllDestinations()
	if err != nil {
		return false, err
	}

	for _, d := range destinations {
		if d.Manufacturer() == manufactuer && d.Name() == name {
			return true, nil
		}
	}
	return false, nil
}
