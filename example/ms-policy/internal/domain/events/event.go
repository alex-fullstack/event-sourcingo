package events

import (
	"policy/internal/domain/dto"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type PolicyEventType int

const (
	RoleCreated PolicyEventType = iota + 1
	PermissionCreated
	UserAssigned
)

func (uet PolicyEventType) String() string {
	return [...]string{
		"role_created",
		"permission_created",
		"user_assigned",
	}[uet-1]
}

func (uet PolicyEventType) Index() int {
	return int(uet)
}

type IntegrationEventPayload struct {
	Role        *dto.RoleProjection        `json:"role"`
	Permissions []dto.PermissionProjection `json:"permissions"`
	Users       []uuid.UUID                `json:"users"`
}

func NewIntegrationEvent(evType int, projection dto.PolicyProjection) (events.IntegrationEvent, error) {
	return events.NewIntegrationEvent(
		projection.ID,
		evType,
		IntegrationEventPayload{
			Role:        projection.Role,
			Permissions: projection.Permissions,
			Users:       projection.Users,
		},
	)
}
