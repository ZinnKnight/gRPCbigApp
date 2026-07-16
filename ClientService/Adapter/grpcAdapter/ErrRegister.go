package grpcAdapter

import (
	"gRPCbigapp/ClientService/Domain"
	"gRPCbigapp/Shared/ErrorInterceptor"
)

func init() {
	ErrorInterceptor.RegisterError(Domain.ErrIncorrectCredentials, ErrorInterceptor.Unauthenticated, "Некорректный логин и пароль")
	ErrorInterceptor.RegisterError(Domain.ErrEmptyName, ErrorInterceptor.Invalid, "Имя не может быть пустым")
	ErrorInterceptor.RegisterError(Domain.ErrEmptyPassword, ErrorInterceptor.Invalid, "Пароль не может быть пустым")
	ErrorInterceptor.RegisterError(Domain.ErrUserNotFound, ErrorInterceptor.NotFound, "Такого пользователя не существует")
	ErrorInterceptor.RegisterError(Domain.ErrUserAlreadyExists, ErrorInterceptor.AlreadyExists, "Такой пользователь уже существует")
	ErrorInterceptor.RegisterError(Domain.ErrTooManyLoginAttempts, ErrorInterceptor.RateLimited, "Достигнут лимит попыток входа")
}
