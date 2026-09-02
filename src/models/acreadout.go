package models

// Readings holds one measured quantity for all phases. A phase that was not
// measured simply has no entry, which reads back as zero but can still be told
// apart from a measured zero by looking up the key.
type Readings map[SmartPiPhase]float64
