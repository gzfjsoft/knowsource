package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/constants"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传原始文档，如果zip,自动解压
func UploadRawDocumentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		err := r.ParseMultipartForm(constants.KnowledgeMaxUploadSize)
		if err != nil {
			httpx.OkJsonCtx(ctx, w, response.Fail(response.InvalidRequestParamCode, "解析表单失败:"+err.Error()))
			return
		}

		documentCode := r.FormValue("documentCode") // 文档类型
		fileType := r.FormValue("fileType")         // 文件类型
		tag := r.FormValue("tag")                   // 标签

		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.OkJsonCtx(ctx, w, response.Fail(response.InvalidRequestParamCode, "文件上传失败:"+err.Error()))
			return
		}
		defer file.Close()

		req := types.UploadRawDocumentsRequest{
			FileName:     header.Filename,
			FileType:     fileType,
			DocumentCode: documentCode,
			Tag:          tag,
		}

		l := knowdata.NewUploadRawDocumentsLogic(ctx, svcCtx)
		resp, err := l.UploadRawDocuments(&req, &file)

		utils.WriteResponse(w, r, err, "上传原始文档", resp)

	}
}
