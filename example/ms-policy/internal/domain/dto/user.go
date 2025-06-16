package dto

type UserEventType int

const (
	CredentialsCreated UserEventType = iota + 1
	UserSigned
	UserAuthenticated
	ConfirmationCreated
	ConfirmationUpdated
	EmailConfirmed
	UserConfirmationRequested
	UserConfirmationRejected
	UserConfirmationCompleted
)

func (uet UserEventType) Index() int {
	return int(uet)
}

type UserAssign struct {
	Email string
	Role  string
}

type UserAssignInput struct {
	ID string `json:"id"`
}

func NewUserAssign(email string, role string) UserAssign {
	return UserAssign{
		Email: email,
		Role:  role,
	}
}

func NewUserAssignInput(id string) UserAssignInput {
	return UserAssignInput{
		ID: id,
	}
}
