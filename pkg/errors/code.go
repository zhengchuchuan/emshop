package errors

import (
    "fmt"
    "net/http"
    "sync"

    gstatus "google.golang.org/grpc/status"
    gcodes "google.golang.org/grpc/codes"
)

var (
	unknownCoder defaultCoder = defaultCoder{1, http.StatusInternalServerError, "An internal server error occurred", "http://emshop/pkg/errors/README.md"}
)

// Coder defines an interface for an error code detail information.
type Coder interface {
	// HTTP status that should be used for the associated error code.
	HTTPStatus() int

	// External (user) facing error text.
	String() string

	// Reference returns the detail documents for user.
	Reference() string

	// Code returns the code of the coder
	Code() int
}

type defaultCoder struct {
	// C refers to the integer code of the ErrCode.
	C int

	// HTTP status that should be used for the associated error code.
	HTTP int

	// External (user) facing error text.
	Ext string

	// Ref specify the reference document.
	Ref string
}

// Code returns the integer code of the coder.
func (coder defaultCoder) Code() int {
	return coder.C

}

// String implements stringer. String returns the external error message,
// if any.
func (coder defaultCoder) String() string {
	return coder.Ext
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
func (coder defaultCoder) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}

	return coder.HTTP
}

// Reference returns the reference document.
func (coder defaultCoder) Reference() string {
	return coder.Ref
}

// codes contains a map of error codes to metadata.
var codes = map[int]Coder{}
var codeMux = &sync.Mutex{}

// Register register a user define error code.
// It will overrid the exist code.
func Register(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved by `emshop/pkg/errors` as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code '0' is reserved by 'emshop/pkg/errors' as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
    if err == nil {
        return nil
    }

    // If this is a gRPC status error, map to HTTP directly
    if st, ok := gstatus.FromError(err); ok {
        return grpcStatusToCoder(st)
    }

    if v, ok := err.(*withCode); ok {
        if coder, ok := codes[v.code]; ok {
            return coder
        }

        // If withCode contains a gRPC status code that we didn't register, map it
        if st, ok := gstatus.FromError(v.err); ok {
            return grpcStatusToCoder(st)
        }
    }

    return unknownCoder
}

// grpcStatusToCoder converts a gRPC status to a defaultCoder with appropriate HTTP status.
func grpcStatusToCoder(st *gstatus.Status) Coder {
    httpStatus := http.StatusInternalServerError
    switch st.Code() {
    case gcodes.OK:
        httpStatus = http.StatusOK
    case gcodes.Canceled:
        httpStatus = http.StatusBadRequest
    case gcodes.Unknown:
        httpStatus = http.StatusInternalServerError
    case gcodes.InvalidArgument:
        httpStatus = http.StatusBadRequest
    case gcodes.DeadlineExceeded:
        httpStatus = http.StatusGatewayTimeout
    case gcodes.NotFound:
        httpStatus = http.StatusNotFound
    case gcodes.AlreadyExists:
        httpStatus = http.StatusConflict
    case gcodes.PermissionDenied:
        httpStatus = http.StatusForbidden
    case gcodes.Unauthenticated:
        httpStatus = http.StatusUnauthorized
    case gcodes.ResourceExhausted:
        httpStatus = http.StatusTooManyRequests
    case gcodes.FailedPrecondition:
        httpStatus = http.StatusBadRequest
    case gcodes.Aborted:
        httpStatus = http.StatusConflict
    case gcodes.OutOfRange:
        httpStatus = http.StatusBadRequest
    case gcodes.Unimplemented:
        httpStatus = http.StatusNotImplemented
    case gcodes.Internal:
        httpStatus = http.StatusInternalServerError
    case gcodes.Unavailable:
        httpStatus = http.StatusServiceUnavailable
    case gcodes.DataLoss:
        httpStatus = http.StatusInternalServerError
    }

    return defaultCoder{C: int(st.Code()), HTTP: httpStatus, Ext: st.Message(), Ref: ""}
}

// IsCode reports whether any error in err's chain contains the given error code.
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}

		if v.cause != nil {
			return IsCode(v.cause, code)
		}

		return false
	}

	return false
}

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
