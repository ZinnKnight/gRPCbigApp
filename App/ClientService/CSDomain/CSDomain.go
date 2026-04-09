package CSDomain

import "errors"

const (
	Free      UserPlan = "FREE_PLAN_USER"
	Pro       UserPlan = "PRO_PLAN_USER"
	AdminRole UserPlan = "ADMIN"
)

type User struct {
	UserID       string
	UserName     string
	UserPassword string
	UserRole     UserPlan
}

type UserPlan string

type IsAdmin struct {
	UserID string
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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrEmptyName         = errors.New("user name is empty")
	ErrEmptyPassword     = errors.New("user password is empty")
)
