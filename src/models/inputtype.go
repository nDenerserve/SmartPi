package models

type InputType int

const (
	NotUsed = iota
	Voltage
	Current
)

func (i InputType) String() string {
	return [...]string{"NotUsed", "Voltage", "Current"}[i]
}

func (i InputType) UnitSymbol() string {
	return [...]string{"N/A", "V", "A"}[i]
}

func InputTypesAsStrings(input []InputType) []string {
	s := make([]string, len(input))
	for i, v := range input {
		s[i] = v.String()
	}
	return s
}
