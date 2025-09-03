package commands

import (
	"user/internal/domain/dto"
	"user/internal/domain/events"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
)

type AuthCommandType int

const (
	SignCmd AuthCommandType = iota + 1
	LoginCmd
	VerificationCmd
)

func (uct AuthCommandType) String() string {
	return [...]string{"sign", "login"}[uct-1]
}

func (uct AuthCommandType) Index() int {
	return int(uct)
}

func NewSignCommand(
	credentials dto.CredentialsCreateInput,
	activity dto.ActivityInput,
) commands.Command {
	es := []commands.CommandEvent{
		commands.NewCommandEvent(events.CredentialsCreated.Index(), credentials),
		commands.NewCommandEvent(events.UserSigned.Index(), activity),
	}
	return commands.NewCommand(SignCmd.Index(), es)
}

func NewLoginCommand(data dto.ActivityInput) commands.Command {
	es := []commands.CommandEvent{
		commands.NewCommandEvent(events.UserAuthenticated.Index(), data),
	}
	return commands.NewCommand(LoginCmd.Index(), es)
}

func NewRequestConfirmCommand(
	confirmation dto.ConfirmationInput,
	activity dto.ActivityInput,
	isNew bool,
) commands.Command {
	var es []commands.CommandEvent
	if isNew {
		es = append(es, commands.NewCommandEvent(events.ConfirmationCreated.Index(), confirmation))
	} else {
		es = append(es, commands.NewCommandEvent(events.ConfirmationUpdated.Index(), confirmation))
	}
	es = append(es, commands.NewCommandEvent(events.UserConfirmationRequested.Index(), activity))
	return commands.NewCommand(VerificationCmd.Index(), es)
}

func NewConfirmCommand(
	activity dto.ActivityInput,
	hasError bool,
) commands.Command {
	var es []commands.CommandEvent
	if hasError {
		es = append(es, commands.NewCommandEvent(events.UserConfirmationRejected.Index(), activity))
	} else {
		es = append(es, commands.NewCommandEvent(events.EmailConfirmed.Index(), struct{}{}))
		es = append(es, commands.NewCommandEvent(events.UserConfirmationCompleted.Index(), activity))
	}

	return commands.NewCommand(VerificationCmd.Index(), es)
}
