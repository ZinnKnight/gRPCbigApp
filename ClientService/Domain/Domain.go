package Domain

import (
	"errors"
	"gRPCbigapp/ClientService/Auth/AuthRoles"
)

const (
	Free      UserPlan = UserPlan(AuthRoles.Free)
	Pro       UserPlan = UserPlan(AuthRoles.Pro)
	AdminRole UserPlan = UserPlan(AuthRoles.Admin)
)

type User struct {
	UserID       string
	UserName     string
	UserPassword string
	UserRole     UserPlan
}

type UserPlan string

type IsAdmin struct {
	UserName string
}

func CanSelfPlanChange(plan UserPlan) bool {
	return plan == Free || plan == Pro
}

func (usr *User) ValidateUser() error {
	if usr.UserName == "" {
		return ErrEmptyName
	}
	if usr.UserPassword == "" {
		return ErrEmptyPassword
	}
	return nil
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrEmptyName            = errors.New("user name is empty")
	ErrEmptyPassword        = errors.New("user password is empty")
	ErrIncorrectCredentials = errors.New("incorrect credentials")
	ErrTooManyLoginAttempts = errors.New("too many loginAttempts")
)
