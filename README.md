# MIDI Router

## Introduction

MIDI Router is an advanced MIDI routing system for MacOS.

In short, MIDIRouter is able to:

  - Watch for MIDI messages on an input interface
  - Replay or not these eventually transformed messages on an output MIDI interface

__Mmhh-kay.. can you give me some examples?"__

The most simple use is to replay all messages from the input interface to the output interface.

Now you may simply want to replay all messages on MIDI channel 1 to the output interface (and ignore others).

But maybe you want to replay all those MIDI messages on channel 1 to the output interface but on channel 5 (and ignore others).

__Hey, but can I transform a Pitch bend event to an Aftertouch event?"__

Yes, you can!

__Hey, but can I transform a Control Change message with number 4 received on Channel 7 to a Sysex messages, with value encoded into 14 bits?"__

Yes, you can!

Well, now you got the idea :)

## Licensing

MIDIRouter is __free for personal use__ (artists, hobbyists, just-want-to-try-ists).

For __commercial and professional use__ I'll ask for a 50€ fee that will give you access to:

  - Pre-Compiled binaries
  - Software support
  - Configuration support

For licensing questions, contact me at __midirouter [at] radix-studio.fr__

## Future work

This program is still under construction, so you may encounter bugs, changes, etc.

I plan to:

  - Support CCAh output messages
  - Add a few commented configuration examples
  - Allow interconnection of two MIDIRouter instances, over an IP network (so yes, you would be able to filter, tranform and re-emit MIDI messages to another MIDI device, somewhere on the Internet :) )

## Tutorials

I'm setting up a set of Use cases with their associated configuration file.

  - ["Simple Forward from a MIDI interface to another"](https://radix-studio.fr/blog/2021/04/16/midirouter-by-example/)
  - ["Note On / Note Off forward only"](https://radix-studio.fr/blog/2021/04/16/midirouter-by-example/)
  - ["Transform a Pitch Bend to an Aftertouch event (and change MIDI channel)"](https://radix-studio.fr/blog/2021/04/16/midirouter-by-example-part2/)
  - "Using notes to emit Program Change events" (soon)
  - "Generate a Sysex message from a Control Change event" (soon)
  - "Using Transform to change value ranges" (soon)

# Configuration

## General settings:

| Name               | Type    | Description                                     |
| ------------------ | ------- | ----------------------------------------------- |
| SourceDevice       | string  | MIDI input device                               |
| DestinationDevice  | string  | MIDI output device                              |
| DefaultPassthrough | bool    | When no filter matches, replay packet "as it"   |
| SendLimitMs        | integer | Limit number of output MIDI messages per second |

## Rules settings:

All filters are declared in a "Rules" JSON array and processed on the configuration file order.

A Rule is built using three items:

  - A filter, used to match a MIDI message read from the MIDI input device
  - A Transformation, used to optionally modify the matched MIDI messaged
  - A Generator, used to create and play a MIDI message on the MIDI output device

### Filters

Filter description depends on the Filter Type (Program Change, Note On/Off, CC, etc.) but all of them share some parameters:

| Name     | Type   | Description                                     |
| -------- | ------ | ----------------------------------------------- |
| Name     | string | A human readable string, describing this filter |
| Channel  | string | The MIDI channel to match (1-16 or *)           |
| MsgType  | string | The type of midi message to match (see below)   |
| Settings | object | Message Type specfic settings (see below)       |

The following message types (MsgType) can be used:

  - Note On
  - Note Off
  - Aftertouch
  - Control Change
  - Program Change
  - Channel Pressure
  - Pitch Wheel
  - *

#### Note On settings

| Name     | Type                               | Description                                     |
| -------- | ---------------------------------- | ----------------------------------------------- |
| Note     | Integer value between 00 and 127   | Note number (Middle C is 60). Use "*" for any   |
| Velocity | Integer value between 00 and 127   | Velocity value. Use "*" for any                 |

#### Note Off settings

| Name     | Type                               | Description                                     |
| -------- | ---------------------------------- | ----------------------------------------------- |
| Note     | Integer value between 00 and 127   | Note number (Middle C is 60). Use "*" for any   |
| Velocity | Integer value between 00 and 127   | Velocity value. Use "*" for any                 |

#### Aftertouch settings

| Name     | Type                               | Description                                     |
| -------- | ---------------------------------- | ----------------------------------------------- |
| Pressure | Integer value between 00 and 127   | Pressure value. Use "*" for any                 |

#### Control Change settings

| Name             | Type                               | Description                                     |
| ---------------- | ---------------------------------- | ----------------------------------------------- |
| Mode             | String                             | Describe how many CC are sent  (see below)      |
| ControllerNumber | Integer value between 00 and (127 or 31)   | Controller Number value. Use "*" for any        |
| Value            | Integer value between 00 and 127   | Control value. Use "*" for any                  |

The following modes are implemented:

  - Standard: 1 Control Change message, value is 7 bits (controller number is from 0 to 127)
  - CCAh : 2 Control Change message, value is 14 bits (controller number is from 0 to 31)

On CCAh (Faderfox EC4) each CC is coded into 2 CC messages:

  - Controler number goes from 0 to 31 (5 bits)
  - On first message, ControllerNumber is the controller numer as "it" (from 0 to 31)
  - On second message, ControllerNumber is the controller number + 0x20



#### Program Change settings

| Name             | Type                               | Description                             |
| ---------------- | ---------------------------------- | --------------------------------------- |
| ProgramNumber    | Integer value between 00 and 127   | Program number. Use "*" for any         |


#### Channel Pressure settings

| Name             | Type                               | Description                             |
| ---------------- | ---------------------------------- | --------------------------------------- |
| Pressure         | Integer value between 00 and 127   | Pressure value. Use "*" for any         |

#### Pitch Wheel settings

| Name             | Type                               | Description                             |
| ---------------- | ---------------------------------- | --------------------------------------- |
| Pitch            | Integer value between 00 and 127   | Pitch value. Use "*" for any         |



### Transformations

Transformations are optional and if not specified, no transformation will be applied to the value extracted by the filter.
They will typically be used to convert a value to the one displayed on a Controller.

A transform will contain the following items:

| Name                  | Description                                                                                                 |
| --------------------- | ------------------------------------------------------------------------------------------------------------|
| Mode                  | "None": No transformation. "Linear": liear scale. "LinearDrop": linear scale & drop out of range values.    |
| FromMin               | Minimal expected value to be received on input                                                              |
| FromMax               | Maximum value to be received on input                                                                       |
| ToMin                 | Minimal value to be generated                                                                               |
| ToMax                 | Maximum value to be generated                                                                               |

When using "Linear" mode, transformation will transpose a value from [FromMin, FromMax] to a value [ToMin, ToMax] using a simple linear extrapolation.
The "LinearDrop" mode will do the same, but drop all input values out of [FromMin, FromMax] and computed output value out of ToMin, ToMax].

__Example:__

Let's say your MIDI controller is used to set a % value from 0 to 100. Actually your destination device expects a value from 0 to 127.
You will use the following configuration:

    "Transform": {
      "Mode": "Linear",
      "FromMin": 0,
      "FromMax": 100,
      "ToMin": 0,
      "ToMax": 127
    }

Now imagine your have the same setup but .. well.. your controller lets you pick a value from 101 to 127 (which is pretty weird forf a %). So you just want to ignore values from 101 to 127.

    "TransformDrop": {
      "Mode": "Linear",
      "FromMin": 0,
      "FromMax": 100,
      "ToMin": 0,
      "ToMax": 127
    }

### Generator

Generator settings depends on the Message Type (Program Change, Note On/Off, CC, etc.) but all of them share some parameters:

| Name     | Type   | Description                                        |
| -------- | ------ | -------------------------------------------------- |
| Name     | string | A human readable string, describing this generator |
| MsgType  | string | The type of generated midi message (see below)     |
| Channel  | string | The MIDI channel to use (1-16 or *)                |
| Settings | object | Message Type specfic settings (see below)          |

The following message types (MsgType) can be used:

  - Note On
  - Note Off
  - Aftertouch
  - Control Change
  - Program Change
  - Channel Pressure
  - Pitch Wheel

When using "*" as MIDI channel, the MIDI channel of the filtered message is re-used.

Each message type has its own settings, but the following special values can be used:

  - "*" : reuse original value taken from filtered message
  - "$" : use value extracted by the filter

#### Note On settings

| Name     | Type                               | Description                    |
| -------- | ---------------------------------- | ------------------------------ |
| Note     | Integer value between 00 and 127   | Note number (Middle C is 60).  |
| Velocity | Integer value between 00 and 127   | Velocity value.                |

The following values can also be set:

  - * : use the original value. Will only be valid if filter is a NoteOn of NoteOff message.
  - $ : use the extracted value by the filter

