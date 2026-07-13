package presentation

type SparkState struct {
	Indicator string
	Label     string
	Role      Role
}

func (s SparkState) Text() string {
	return s.Indicator + " " + s.Label
}

func StateFor(done, important bool) SparkState {
	switch {
	case done:
		return SparkState{Indicator: "[x]", Label: "done", Role: Completed}
	case important:
		return SparkState{Indicator: "[!]", Label: "important", Role: Important}
	default:
		return SparkState{Indicator: "[ ]", Label: "active", Role: Muted}
	}
}
