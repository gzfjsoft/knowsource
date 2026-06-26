package response

import "github.com/zeromicro/go-zero/core/logx"

//httpxx.ErrorCtx(r.Context(), w, err)
//httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call" , err.Error()))

const (
	SuccessCode                             = 200
	UnauthorizedCode                        = 401
	ForbiddenCode                           = 403
	NotFoundCode                            = 404
	WrongCaptchaCode                        = 451
	UserNotExistCode                        = 452
	RecordNotExistCode                      = 453
	ParameterErrorCode                      = 454
	InvalidRequestParamCode                 = 455
	InvalidRequestParamCodeInHandler        = 499
	MethodNotAllowedCode                    = 405
	ConflictCode                            = 409
	RequestTimeoutCode                      = 408
	GoneCode                                = 410
	PreconditionFailedCode                  = 412
	UnsupportedMediaTypeCode                = 415
	InsufficientBalanceCode                 = 456
	UserBalanceNotExistCode                 = 457
	EmailNotVerifiedCode                    = 458
	ServerErrorCode                         = 500
	InternalServerErrorDuringProcessingCode = 503
)

/*
成功相关状态码

SuccessCode = 200：表示请求成功，操作顺利完成，符合常规的成功响应约定，方便客户端识别请求已被正确处理。

客户端错误相关状态码

UnauthorizedCode = 401：客户端尝试访问需要认证的资源但未提供有效的认证凭据，用于明确权限认证缺失的情况。
ForbiddenCode = 403：客户端有访问请求但被服务器端明确禁止访问，即便提供了认证信息也无权访问特定资源，体现了权限验证后访问被拒的情况，与 401 有所区分。
WrongCaptchaCode = 451：专门用于表示验证码错误的情况，在涉及验证码验证环节时，方便客户端准确判断是验证码相关的验证失败。
UserNotExistCode = 452：当在用户查找等相关操作中找不到对应的用户记录时返回此状态码，有助于客户端针对性地进行后续提示或处理逻辑。
RecordNotExistCode = 453：指示除用户记录外其他业务记录不存在的情况，利于前端等调用方准确判断是具体哪种记录缺失的问题。
ParameterErrorCode = 454：表示传入的参数整体出现错误，相对笼统地指出参数方面存在问题，后续可结合日志等排查具体错误详情。
InvalidRequestParamCode = 455：更强调请求参数不符合要求、格式不对或者不符合接口定义等情况，对参数相关错误进一步细化区分，方便客户端按不同参数问题做不同处理。
MethodNotAllowedCode = 405：当客户端使用了不被允许的 HTTP 方法（比如对某个只支持 GET 的接口使用了 POST 方法）时返回该状态码，用于反馈请求方法层面的问题，规范客户端的接口调用行为。
ConflictCode = 409：常用于在创建资源等操作时，出现了和现有资源冲突的情况，例如试图创建一个已经存在的唯一用户名等场景，便于客户端知晓冲突并采取相应处理措施。
RequestTimeoutCode = 408：如果客户端的请求在服务器端等待处理的时间过长，超过了规定的超时时间，使用此状态码告知客户端其请求超时了，促使客户端考虑优化网络、重试等操作。
GoneCode = 410：表示请求的资源已经被永久删除，以后也不会再得到，适用于体现资源彻底消失的状态，区别于一般的记录不存在情况。
PreconditionFailedCode = 412：当客户端发送的请求头中包含一些前置条件（如请求的资源版本等）不符合服务器端预期时，返回该状态码告知客户端前置条件不满足，引导客户端调整请求条件。
UnsupportedMediaTypeCode = 415：若客户端提交的数据格式不被服务器端支持（比如接口要求 JSON 格式但客户端发送了 XML 格式数据），用这个状态码清晰指出是媒体类型不兼容的问题，便于客户端进行数据格式转换等处理。

业务逻辑错误相关状态码

InsufficientBalanceCode = 456：针对余额不足这种特定业务场景定义的状态码，在涉及支付、消费等业务逻辑时，能让客户端快速知晓是因为余额不够导致操作无法进行，利于进行相应提示给用户。
UserBalanceNotExistCode = 457：表示用户余额不存在，与 InsufficientBalanceCode 相呼应，从不同角度界定了和用户余额相关的异常情况，方便客户端清楚区分是余额不存在还是仅仅余额不足的问题。
EmailNotVerifiedCode = 458：表示租户尚未完成邮箱验证，需要先完成邮箱验证才能登录。

服务器错误相关状态码

ServerErrorCode = 500：遵循类似 HTTP 状态码中 500 代表服务器内部错误的概念，用于告知客户端请求处理过程中服务器端出现了未知的、内部的故障，方便排查是服务端自身的问题而非客户端请求的问题。
InternalServerErrorDuringProcessingCode = 503：用于表示服务器当前暂时无法处理请求，比如正在进行维护、过载等情况，让客户端知道可以稍后重试，比单纯的 500 能传递更具体的服务不可用场景信息。
*/
type Response struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Info    string      `json:"info,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

//go:inline
func OK(data interface{}) Response {
	return Success("ok", data)
}

//go:inline
func Success(message string, data interface{}) Response {
	return Response{Code: SuccessCode, Message: message, Data: data}
}

//go:inline
func Error(message string) Response {
	return FailWithInfo(ServerErrorCode, message, "")
}

//go:inline
func ErrorWithInfo(message string, info string) Response {
	return FailWithInfo(ServerErrorCode, message, info)
}

//go:inline
func Fail(code int64, message string) Response {
	return FailWithInfo(code, message, "")
}

//go:inline
func FailWithInfo(code int64, message string, info string) Response {
	if message == "" {
		message = "server error"
	}
	logx.Errorf("code: %+v, message: %+v, info: %+v", code, message, info)
	return Response{Code: code, Message: message, Info: info}
}
