package entities

import (
	"errors"
	"policy/internal/domain/dto"
	"policy/internal/domain/events"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	coreEvents "gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type Policy struct {
	*entities.Aggregate
	role        *Role
	permissions []*Permission
	users       []uuid.UUID
}

func NewPolicy(id uuid.UUID) *Policy {
	res := &Policy{
		permissions: make([]*Permission, 0),
		users:       make([]uuid.UUID, 0),
	}
	res.Aggregate = entities.NewAggregate(id, func(event coreEvents.Event) error {
		payload := event.Payload
		var err error
		var user = uuid.Nil
		var permission *Permission
		switch events.PolicyEventType(event.Type) {
		case events.RoleCreated:
			res.role, err = NewRole(payload)
		case events.PermissionCreated:
			permission, err = NewPermission(payload)
		case events.UserAssigned:
			userID, ok := payload["id"].(string)
			if !ok {
				err = errors.New("invalid user assigned payload")
			} else {
				user, err = uuid.Parse(userID)
			}
		default:
			err = errors.New("unhandled default case")
		}
		if err != nil {
			return err
		}
		if permission != nil {
			res.permissions = append(res.permissions, permission)
		}
		if user != uuid.Nil {
			res.users = append(res.users, user)
		}
		return nil
	})
	return res
}

func (p *Policy) Projection() interface{} {
	permissions := make([]dto.PermissionProjection, len(p.permissions))
	for i, perm := range p.permissions {
		permissions[i] = dto.NewPermissionProjection(perm.Code, perm.Name)
	}
	var role *dto.RoleProjection
	if p.role != nil {
		role = dto.NewRoleProjection(p.role.Code.String(), p.role.Name)
	}

	return dto.PolicyProjection{
		ID:          p.ID(),
		Role:        role,
		Permissions: permissions,
		Users:       p.users,
	}
}

func (p *Policy) IntegrationEvent(evType int) (coreEvents.IntegrationEvent, error) {
	projection, ok := p.Projection().(dto.PolicyProjection)
	if !ok {
		return coreEvents.IntegrationEvent{}, errors.New("wrong data type, need dto.PolicyProjection")
	}
	return events.NewIntegrationEvent(evType, projection)
}
