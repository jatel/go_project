package innererror

import "fmt"

type ErrCode int

const (
	ErrNoCode           ErrCode = -2
	ErrNoError          ErrCode = 0
	ErrUnknown          ErrCode = -1
	ErrSQLError         ErrCode = 45002
	ErrDecodeError      ErrCode = 45003
	ErrInvalidParaError ErrCode = 45004
	ErrOutOfRangeError  ErrCode = 45005
	ErrRPCCallError     ErrCode = 45006
)

func (err ErrCode) ErrorInfo() string {
	switch err {
	case ErrNoCode:
		return "success"
	case ErrNoError:
		return "not an error"
	case ErrUnknown:
		return "unknown error"
	case ErrSQLError:
		return "Execute SQL statement to report error"
	case ErrDecodeError:
		return "Decoding request body error"
	case ErrInvalidParaError:
		return "Invalid input parameter"
	case ErrOutOfRangeError:
		return "out of range"
	case ErrRPCCallError:
		return "call json rpc error"
	}

	return fmt.Sprintf("Unknown error? Error code = %d", err)
}

func (err ErrCode) Error() string {
	return err.ErrorInfo()
}

func (err ErrCode) Value() int {
	return int(err)
}
