package ErrorInterceptor

import (
	clientErr "gRPCbigapp/ClientService/Domain"
	spotIntsrumentErr "gRPCbigapp/SpotInstrumentService/Domain"
)

func init() {
	RegisterError(clientErr.ErrIncorrectCredentials, Unauthenticated, "Некорректный логин или пароль")
	RegisterError(clientErr.ErrEmptyName, Invalid, "Имя пользователя не может быть пустым")
	RegisterError(clientErr.ErrEmptyPassword, Invalid, "Пароль не может быть пустым")
	RegisterError(clientErr.ErrUserNotFound, NotFound, "Такого пользователя не существует")
	RegisterError(clientErr.ErrUserAlreadyExists, AlreadyExists, "Такой пользователь уже существует")
	RegisterError(clientErr.ErrTooManyLoginAttempts, RateLimited, "Достигнут лимит попыток входа")

	RegisterError(spotIntsrumentErr.ErrMarketNotFound, NotFound, "Невозможно найти такой магазин")
}
