package errors

import (
	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ParseSvcErrToProto(err error) (error, bool) {
	if svcerr.IsBadRequest(err) {
		return CastInvalidRequest(err), false
	}

	if svcerr.IsNotFound(err) {
		return CastNotFound(err), false
	}

	return status.Error(codes.Internal, ""), true
}

func CastInvalidRequest(err error) error {
	return status.Error(codes.InvalidArgument, err.Error())
}

func CastNotFound(err error) error {
	return status.Error(codes.NotFound, err.Error())
}
