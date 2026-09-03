package models

// AnalogInChannel is one channel's reading, given both as a current (mA, for
// a 4-20mA loop) and a voltage (V, for a 0-10V input). Which one is
// physically meaningful depends on how that channel's jumper is wired on the
// module - the server has no way to read that jumper, so both are always
// reported and the caller uses whichever matches its own wiring.
type AnalogInChannel struct {
	Current float64 `json:"current"`
	Voltage float64 `json:"voltage"`
}

// AnalogInStatus is the reading of one 4-channel analog-in module. Channels
// is indexed 0-3 for physical channels 1-4.
type AnalogInStatus struct {
	Moduleaddress string             `json:"moduleaddress"`
	Channels      [4]AnalogInChannel `json:"channels"`
}

// AnalogInChannelStatus is the reading of a single channel of an analog-in
// module, as returned by the single-channel endpoint - deliberately not
// AnalogInStatus with the other three channels left zeroed, which would be
// indistinguishable from those channels genuinely reading zero.
type AnalogInChannelStatus struct {
	Moduleaddress string `json:"moduleaddress"`
	// Channel is 1-4, the physical labelling on the module.
	Channel int `json:"channel"`
	AnalogInChannel
}
