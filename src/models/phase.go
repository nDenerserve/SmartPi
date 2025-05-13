package models

const (
	_ = iota
	PhaseA
	PhaseB
	PhaseC
	PhaseN
)

type SmartPiPhase uint

func (p SmartPiPhase) String() string {
	switch p {
	case PhaseA:
		return "A"
	case PhaseB:
		return "B"
	case PhaseC:
		return "C"
	case PhaseN:
		return "N"
	}
	panic("Unreachable")
}

func (p SmartPiPhase) PhaseNumber() string {
	switch p {
	case PhaseA:
		return "1"
	case PhaseB:
		return "2"
	case PhaseC:
		return "3"
	case PhaseN:
		return "4"
	}
	panic("Unreachable")
}

func PhaseNameFromNumber(p string) SmartPiPhase {
	switch p {
	case "1":
		return PhaseA
	case "2":
		return PhaseB
	case "3":
		return PhaseC
	case "4":
		return PhaseN
	}
	panic("Unreachable")
}
