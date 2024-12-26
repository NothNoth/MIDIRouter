package main

import (
	"MIDIRouter/config"
	"fmt"
	"os"

	"github.com/youpy/go-coremidi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<config file>")
		fmt.Println("MIDI inputs:")
		sources, err := coremidi.AllSources()
		if err != nil {
			panic(err)
		}
		for _, source := range sources {
			fmt.Println("  " + source.Entity().Device().Name() + "/" + source.Manufacturer() + "/" + source.Name())
		}

		fmt.Println("MIDI outputs:")
		destinations, err := coremidi.AllDestinations()
		if err != nil {
			panic(err)
		}
		for _, destination := range destinations {
			fmt.Println("  " + destination.Manufacturer() + "/" + destination.Name())
		}

		return
	}

	router, err := config.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	router.Start()

}
