package router

import (
	"MIDIRouter/rule"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/youpy/go-coremidi"
)

type MIDIRouter struct {
	sourceDevice      string
	destinationDevice string

	midiClient coremidi.Client
	srcPort    coremidi.InputPort

	destPort    coremidi.OutputPort
	destination coremidi.Destination

	defaultPassThrough bool
	lastMIDIMsg        time.Time
	sendLimit          time.Duration
	rules              []*rule.Rule

	verbose bool
}

func New(sourceDevice string, destinationDevice string) (*MIDIRouter, error) {
	var relay MIDIRouter
	var err error

	relay.sourceDevice = sourceDevice
	relay.destinationDevice = destinationDevice
	relay.defaultPassThrough = false

	relay.midiClient, err = coremidi.NewClient("MIDIRouter")
	if err != nil {
		return nil, err
	}
	err = relay.setupSource()
	if err != nil {
		return nil, err
	}

	err = relay.setupDestination()
	if err != nil {
		return nil, err
	}
	return &relay, nil
}

func (relay *MIDIRouter) SetVerbose(verb bool) {
	relay.verbose = verb
}

func (relay *MIDIRouter) SetPassthrough(pass bool) {
	relay.defaultPassThrough = pass
}

func (relay *MIDIRouter) SetSendLimit(delay time.Duration) {
	relay.sendLimit = delay
}

func (relay *MIDIRouter) Start() {

	for {
		time.Sleep(1 * time.Millisecond)
	}
}

func (relay *MIDIRouter) AddRule(rule *rule.Rule) {
	relay.rules = append(relay.rules, rule)
	fmt.Println(rule)
}

func (relay *MIDIRouter) onPacket(source coremidi.Source, packet coremidi.Packet) {

	if relay.verbose == true {
		fmt.Printf(
			"device: %v, manufacturer: %v, source: %v, data: %v\n",
			source.Entity().Device().Name(),
			source.Manufacturer(),
			source.Name(),
			hex.EncodeToString(packet.Data),
		)
	}

	ruleMAtched := false
	for _, r := range relay.rules {
		if len(packet.Data) == 0 {
			continue
		}

		//Stop on first rule success
		match, newPacket := r.Match(packet, relay.verbose)
		if match == rule.RuleMatchResultMatchInject {
			if relay.verbose {
				fmt.Println("-> Sending generated packet :")
				fmt.Println(hex.Dump(newPacket.Data))
			}

			if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
				fmt.Println("Ignoring midi message (send limit)")
				return
			}
			newPacket.Send(&relay.destPort, &relay.destination)
			relay.lastMIDIMsg = time.Now()
			ruleMAtched = true
			break
		} else if match == rule.RuleMatchResultMatchNoInject {
			ruleMAtched = true
			break
		}
	}

	if (ruleMAtched == false) && (relay.verbose == true) {
		fmt.Println("-> No match")
	}

	//no match, apply passthrough if set
	if (ruleMAtched == false) && (relay.defaultPassThrough == true) {
		if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
			fmt.Println("Ignoring midi message (send limit)")
			return
		}
		packet.Send(&relay.destPort, &relay.destination)
		relay.lastMIDIMsg = time.Now()
	}
	return
}
