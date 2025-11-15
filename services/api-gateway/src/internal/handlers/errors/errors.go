package errors

import (
	"net/http"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/svcerr"
)

func ParseSvcErrToResp(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	if svcerr.IsNotFound(err) {
		return http.StatusNotFound, err.Error()
	}

	if svcerr.IsBadRequest(err) {
		return http.StatusBadRequest, err.Error()
	}

	return http.StatusInternalServerError, ""
}
