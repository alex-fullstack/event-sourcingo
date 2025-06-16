package commands

type CommandEvent struct {
	Type    int
	Payload interface{}
}

type Command struct {
	Type   int
	Events []CommandEvent
}

func NewCommandEvent(eType int, payload interface{}) CommandEvent {
	return CommandEvent{Type: eType, Payload: payload}
}

func NewCommand(cType int, events []CommandEvent) Command {
	return Command{Events: events, Type: cType}
}
