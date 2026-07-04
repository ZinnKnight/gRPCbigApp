package ErrorInterceptor

import (
	clientErr "gRPCbigapp/App/ClientService/Domain"
	orderErr "gRPCbigapp/App/OrderService/Domain"
	spotIntsrumentErr "gRPCbigapp/App/SpotInstrumentService/Domain"
)

func init() {
	RegisterError(clientErr.ErrIncorrectCredentials, Unauthenticated, "Некорректный логин или пароль")
	RegisterError(clientErr.ErrEmptyName, Invalid, "Имя пользователя не может быть пустым")
	RegisterError(clientErr.ErrEmptyPassword, Invalid, "Пароль не может быть пустым")
	RegisterError(clientErr.ErrUserNotFound, NotFound, "Такого пользователя не существует")
	RegisterError(clientErr.ErrUserAlreadyExists, AlreadyExists, "Такой пользователь уже существует")
	RegisterError(clientErr.ErrTooManyLoginAttempts, RateLimited, "Достигнут лимит попыток входа")

	RegisterError(orderErr.ErrOrderNotFound, NotFound, "Такого заказа не существует")
	RegisterError(orderErr.ErrOrderQuotaExceeded, RateLimited, "Достигнут лимит по заказам")
	RegisterError(orderErr.ErrInvalidUserID, Invalid, "Некорректный идентификатор пользователя")
	RegisterError(orderErr.ErrInvalidMarketID, Invalid, "Некорректный идентификатор магазина")
	RegisterError(orderErr.ErrInvalidPrice, Invalid, "Некорректная цена заказа")
	RegisterError(orderErr.ErrInvalidAmount, Invalid, "Некорректное количество в заказе")
	RegisterError(orderErr.ErrOrderAlreadyExists, AlreadyExists, "Заказ уже существует")

	RegisterError(spotIntsrumentErr.ErrMarketNotFound, NotFound, "Невозможно найти такой магазин")
}
