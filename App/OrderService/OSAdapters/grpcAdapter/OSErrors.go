package grpcAdapter

import (
	"errors"
	"gRPCbigapp/App/OrderService/OSDomain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthenticated = status.Error(codes.Unauthenticated, "Required authentication")
)

func DomainErrorMapping(err error) error {
	switch {
	case errors.Is(err, OSDomain.ErrOrderNotFound):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, OSDomain.ErrInvalidPrice), errors.Is(err, OSDomain.ErrInvalidAmount), errors.Is(err, OSDomain.ErrInvalidMarketID),
		errors.Is(err, OSDomain.ErrInvalidUserID):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "Internal error ")
	}
}
