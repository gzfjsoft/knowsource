package utils

import (
	"knowsource/common/response"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ParseRequest[T any](w http.ResponseWriter, r *http.Request, v *T) error {
	err := httpx.Parse(r, v)
	if err != nil {
		httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ParameterErrorCode, "解析请求参数失败", err.Error()))
	}
	return err
}

func ParsePathRequest[T any](w http.ResponseWriter, r *http.Request, v *T) error {
	err := httpx.ParsePath(r, v)
	if err != nil {
		httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ParameterErrorCode, "解析请求参数失败", err.Error()))
	}
	return err
}

func WriteResponse(w http.ResponseWriter, r *http.Request, err error, msg string, v interface{}) {
	if msg == "" {
		msg = "请求执行"
	}
	if err != nil {
		httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, msg+"失败:"+err.Error(), err.Error()))
	} else {
		httpx.OkJsonCtx(r.Context(), w, response.Success(msg+"成功", v))
	}
}

// convert string to int64, default -1
func ConvertInt64(v string, defaultValue int64) int64 {
	if v == "" {
		return defaultValue
	}
	result, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	return result
}

func FixPagingParamWithMax(page, pageSize, maxPage uint64) (uint64, uint64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if maxPage > 0 && pageSize > maxPage {
		pageSize = maxPage
	}
	return page, pageSize
}

func FixPagingParam(page, pageSize uint64) (uint64, uint64) {
	return FixPagingParamWithMax(page, pageSize, 100)
}

func CountStrLen(str string) int {
	return len([]rune(str))
}
