package commands

import (
	"policy/internal/domain/dto"
	"policy/internal/domain/events"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
	"github.com/google/uuid"
)

type PolicyCommandType int

const (
	NewUserCmd PolicyCommandType = iota + 1
	NewRoleCmd
	NewPermissionCmd
	NewUserAssignCmd
)

func (uct PolicyCommandType) String() string {
	return [...]string{"new_user", "new_role", "new_permission", "new_user_assign"}[uct-1]
}

func (uct PolicyCommandType) Index() int {
	return int(uct)
}

func NewCreateRoleCommand(create dto.RoleCreate) commands.Command {
	return commands.NewCommand(
		NewRoleCmd.Index(),
		[]commands.CommandEvent{
			commands.NewCommandEvent(
				events.RoleCreated.Index(),
				create,
			),
		},
	)
}

func NewCreatePermissionCommand(create dto.PermissionCreateInput) commands.Command {
	return commands.NewCommand(
		NewPermissionCmd.Index(),
		[]commands.CommandEvent{
			commands.NewCommandEvent(
				events.PermissionCreated.Index(),
				create,
			),
		},
	)
}

func NewUserAssignCommand(userID uuid.UUID) commands.Command {
	return commands.NewCommand(
		NewUserAssignCmd.Index(),
		[]commands.CommandEvent{
			commands.NewCommandEvent(
				events.UserAssigned.Index(),
				dto.NewUserAssignInput(userID.String()),
			),
		},
	)
}
