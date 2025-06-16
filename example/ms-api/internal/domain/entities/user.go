package entities

import (
	"api/internal/domain/dto"
	"slices"
	"time"

	"github.com/google/uuid"
)

type UserHistoryRecord struct {
	ActivityType string
	Timestamp    time.Time
	Device       string
}

func NewUserHistoryRecord(dto dto.UserHistoryRecord) *UserHistoryRecord {
	return &UserHistoryRecord{
		ActivityType: dto.ActivityType,
		Timestamp:    dto.Timestamp,
		Device:       dto.Device,
	}
}

type UserInfo struct {
	Email          string
	ActivationCode *string
	EmailVerified  bool
	History        []*UserHistoryRecord
}

func NewUserInfo(dto dto.UserUpsert) *UserInfo {
	return &UserInfo{
		Email:          dto.Email,
		ActivationCode: dto.ActivationCode,
		EmailVerified:  dto.EmailVerified,
		History: slices.Collect(func(yield func(*UserHistoryRecord) bool) {
			for _, rec := range dto.History {
				if !yield(NewUserHistoryRecord(rec)) {
					return
				}
			}
		}),
	}
}

type User struct {
	id       uuid.UUID
	info     *UserInfo
	policies []*Policy
}

func NewUser(id uuid.UUID, info dto.UserUpsert, policies []dto.UserPolicy) *User {
	u := &User{
		id:       id,
		info:     NewUserInfo(info),
		policies: make([]*Policy, 0),
	}
	if len(policies) > 0 {
		u.policies = slices.Collect(func(yield func(policy *Policy) bool) {
			for _, policy := range policies {
				if !yield(NewPolicy(policy)) {
					return
				}
			}
		})
	}
	return u
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Document() dto.UserDocument {
	history := make([]dto.UserHistoryRecord, 0)
	if len(u.info.History) > 0 {
		history = slices.Collect(func(yield func(record dto.UserHistoryRecord) bool) {
			for _, item := range u.info.History {
				rec := dto.UserHistoryRecord{
					ActivityType: item.ActivityType,
					Timestamp:    item.Timestamp,
					Device:       item.Device,
				}
				if !yield(rec) {
					return
				}
			}
		})
	}
	policies := make([]dto.UserPolicy, 0)
	if len(u.policies) > 0 {
		policies = slices.Collect(func(yield func(policy dto.UserPolicy) bool) {
			for _, policy := range u.policies {
				var role *dto.Role
				doc := policy.Document()
				if doc.Role != nil {
					role = &dto.Role{
						Code: doc.Role.Code,
						Name: doc.Role.Name,
					}
				}
				if !yield(dto.UserPolicy{
					ID:          doc.ID,
					Role:        role,
					Permissions: collectPermissions(doc.Permissions),
				}) {
					return
				}
			}
		})
	}
	return dto.UserDocument{
		ID: u.id,
		Info: dto.UserUpsert{
			Email:          u.info.Email,
			ActivationCode: u.info.ActivationCode,
			EmailVerified:  u.info.EmailVerified,
			History:        history,
		},
		Policies: policies,
	}
}

func collectPermissions(ps []dto.Permission) []dto.Permission {
	permissions := make([]dto.Permission, 0)
	if len(ps) > 0 {
		permissions = slices.Collect(func(yield func(permission dto.Permission) bool) {
			for _, perm := range ps {
				if !yield(
					dto.Permission{
						Code: perm.Code,
						Name: perm.Name,
					}) {
					return
				}
			}
		})
	}
	return permissions
}
