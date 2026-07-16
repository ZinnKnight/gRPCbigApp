package grpcAdapter

import (
	"gRPCbigapp/OrderService/Domain"
	"gRPCbigapp/Shared/ErrorInterceptor"
)

func init() {
	ErrorInterceptor.RegisterError(Domain.ErrOrderNotFound, ErrorInterceptor.NotFound, "Указанного заказа не существует")
	ErrorInterceptor.RegisterError(Domain.ErrOrderQuotaExceeded, ErrorInterceptor.RateLimited, "Исчерпан лимит заказов")
	ErrorInterceptor.RegisterError(Domain.ErrInvalidUserID, ErrorInterceptor.Invalid, "Некорректный идентификатор для пользователя")
	ErrorInterceptor.RegisterError(Domain.ErrInvalidMarketID, ErrorInterceptor.Invalid, "Некорректный идентификатор для магазина")
	ErrorInterceptor.RegisterError(Domain.ErrInvalidPrice, ErrorInterceptor.Invalid, "Некорректная цена")
	ErrorInterceptor.RegisterError(Domain.ErrInvalidAmount, ErrorInterceptor.Invalid, "Некорректное количество")
	ErrorInterceptor.RegisterError(Domain.ErrOrderAlreadyExists, ErrorInterceptor.AlreadyExists, "Такой заказ уже существует")
}
