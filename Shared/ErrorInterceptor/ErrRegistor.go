package ErrorInterceptor

import (
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

func init() {
	RegisterError(CSDomain.ErrIncorrectCredentials, Unauthenticated, "Некорректный логин или пароль")
	RegisterError(CSDomain.ErrEmptyName, Invalid, "Имя пользователя не может быть пустым")
	RegisterError(CSDomain.ErrEmptyPassword, Invalid, "Пароль не может быть пустым")
	RegisterError(CSDomain.ErrUserNotFound, NotFound, "Такого пользователя не существует")
	RegisterError(CSDomain.ErrUserAlreadyExists, AlreadyExists, "Такой пользователь уже существует")

	RegisterError(OSDomain.ErrOrderNotFound, NotFound, "Такого заказа не существует")
	RegisterError(OSDomain.ErrInvalidUserID, Invalid, "Некорректный идентификатор пользователя")
	RegisterError(OSDomain.ErrInvalidMarketID, Invalid, "Некорректный идентификатор магазина")
	RegisterError(OSDomain.ErrInvalidPrice, Invalid, "Некорректная цена заказа")
	RegisterError(OSDomain.ErrInvalidAmount, Invalid, "Некорректное количество в заказе")
	RegisterError(OSDomain.ErrOrderAlreadyExists, AlreadyExists, "Заказ уже существует")

	RegisterError(SISDomain.ErrMarketNotFound, NotFound, "Невозможно найти такой магазин")
}
