package commands

type CommandEvent[T any] struct {
	Type    int
	Payload T
}

type Command[T any] struct {
	Type   int
	Events []CommandEvent[T]
}

func NewCommandEvent[T any](eType int, payload T) CommandEvent[T] {
	return CommandEvent[T]{Type: eType, Payload: payload}
}

func NewCommand[T any](cType int, events []CommandEvent[T]) Command[T] {
	return Command[T]{Events: events, Type: cType}
}
