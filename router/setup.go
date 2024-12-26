package router

import (
	"errors"
	"fmt"

	"github.com/youpy/go-coremidi"
)

func (relay *MIDIRouter) setupSource() error {
	source, err := findSource(relay.sourceDevice)
	if err != nil {
		return err
	}
	relay.srcPort, err = coremidi.NewInputPort(relay.midiClient, relay.sourceDevice+" input port",
		func(source coremidi.Source, packet coremidi.Packet) {
			relay.onPacket(source, packet)
		})
	if err != nil {
		return err
	}

	_, err = relay.srcPort.Connect(source)
	if err != nil {
		return err
	}

	return nil
}

func (relay *MIDIRouter) setupDestination() error {
	destination, err := findDestination(relay.destinationDevice)
	if err != nil {
		return err
	}

	relay.destPort, err = coremidi.NewOutputPort(relay.midiClient, relay.destinationDevice+" output port")
	if err != nil {
		return err
	}

	relay.destination = destination
	fmt.Println("Destination device: ", destination.Name(), "(", destination.Manufacturer(), ")")

	return nil
}

func findSource(key string) (coremidi.Source, error) {
	sources, err := coremidi.AllSources()
	if err != nil {
		return coremidi.Source{}, err
	}

	for _, s := range sources {
		dk := s.Entity().Device().Name() + "/" + s.Manufacturer() + "/" + s.Name()
		if dk == key {
			return s, nil
		}
	}

	return coremidi.Source{}, errors.New("MIDI source not found: " + key)
}

func findDestination(key string) (coremidi.Destination, error) {
	dest, err := coremidi.AllDestinations()
	if err != nil {
		return coremidi.Destination{}, err
	}

	for _, d := range dest {
		dk := d.Manufacturer() + "/" + d.Name()
		if dk == key {
			return d, nil
		}
	}

	return coremidi.Destination{}, errors.New("MIDI destination not found: " + key)
}
