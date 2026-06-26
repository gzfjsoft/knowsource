package types

func TypeResponse(code int64, message string, info string) Response {
	return Response{
		Code:    code,
		Message: message,
		Info:    info,
	}
}
