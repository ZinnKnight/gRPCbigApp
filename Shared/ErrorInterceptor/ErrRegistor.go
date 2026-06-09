package ErrorInterceptor

import (
	"gRPCbigapp/App/ClientService/Domain"
	"gRPCbigapp/App/OrderService/Domain"
	"gRPCbigapp/App/SpotInstrumentService/Domain"
)

func init() {
	RegisterError(Domain.ErrIncorrectCredentials, Unauthenticated, "Некорректный логин или пароль")
	RegisterError(Domain.ErrEmptyName, Invalid, "Имя пользователя не может быть пустым")
	RegisterError(Domain.ErrEmptyPassword, Invalid, "Пароль не может быть пустым")
	RegisterError(Domain.ErrUserNotFound, NotFound, "Такого пользователя не существует")
	RegisterError(Domain.ErrUserAlreadyExists, AlreadyExists, "Такой пользователь уже существует")

	RegisterError(Domain.ErrOrderNotFound, NotFound, "Такого заказа не существует")
	RegisterError(Domain.ErrInvalidUserID, Invalid, "Некорректный идентификатор пользователя")
	RegisterError(Domain.ErrInvalidMarketID, Invalid, "Некорректный идентификатор магазина")
	RegisterError(Domain.ErrInvalidPrice, Invalid, "Некорректная цена заказа")
	RegisterError(Domain.ErrInvalidAmount, Invalid, "Некорректное количество в заказе")
	RegisterError(Domain.ErrOrderAlreadyExists, AlreadyExists, "Заказ уже существует")

	RegisterError(Domain.ErrMarketNotFound, NotFound, "Невозможно найти такой магазин")
}
