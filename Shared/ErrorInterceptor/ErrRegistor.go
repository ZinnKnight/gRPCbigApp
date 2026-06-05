package ErrorInterceptor

import (
	"gRPCbigapp/App/ClientService/Domain"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

func init() {
	RegisterError(Domain.ErrIncorrectCredentials, Unauthenticated, "Некорректный логин или пароль")
	RegisterError(Domain.ErrEmptyName, Invalid, "Имя пользователя не может быть пустым")
	RegisterError(Domain.ErrEmptyPassword, Invalid, "Пароль не может быть пустым")
	RegisterError(Domain.ErrUserNotFound, NotFound, "Такого пользователя не существует")
	RegisterError(Domain.ErrUserAlreadyExists, AlreadyExists, "Такой пользователь уже существует")

	RegisterError(OSDomain.ErrOrderNotFound, NotFound, "Такого заказа не существует")
	RegisterError(OSDomain.ErrInvalidUserID, Invalid, "Некорректный идентификатор пользователя")
	RegisterError(OSDomain.ErrInvalidMarketID, Invalid, "Некорректный идентификатор магазина")
	RegisterError(OSDomain.ErrInvalidPrice, Invalid, "Некорректная цена заказа")
	RegisterError(OSDomain.ErrInvalidAmount, Invalid, "Некорректное количество в заказе")
	RegisterError(OSDomain.ErrOrderAlreadyExists, AlreadyExists, "Заказ уже существует")

	RegisterError(SISDomain.ErrMarketNotFound, NotFound, "Невозможно найти такой магазин")
}
