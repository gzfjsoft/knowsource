package constants

import "strings"

// RawDocumentsStatus* defines the lifecycle status of rows in `raw_documents`.
//
// Keep these values stable because the frontend may display them directly.
const (
	// RawDocumentsStatusUploaded 上传完成、尚未开始或已中断识别（可再次点击「识别文字」）
	RawDocumentsStatusUploaded = "已上传"
	RawDocumentsStatusUploadedExtracting = "已上传，正在提取文字..."
	RawDocumentsStatusExtracting         = "正在提取文字..."
	RawDocumentsStatusExtractedNotInDB   = "已提取文字未审核入库"
	RawDocumentsStatusInserting          = "正在入库..."
	RawDocumentsStatusInserted           = "已经入库"
	RawDocumentsStatusRemoving           = "正在出库..."
	// RawDocumentsStatusExtractFailed 识别失败（可再次点击「识别文字」）
	RawDocumentsStatusExtractFailed = "识别失败"
	// RawDocumentsStatusInsertFailed 审核入库失败（可再次点击「审核」）
	RawDocumentsStatusInsertFailed = "入库失败"
)

// IsRawDocumentFailedStatus 是否处于失败态（可重试）
func IsRawDocumentFailedStatus(status string) bool {
	s := strings.TrimSpace(status)
	return s == RawDocumentsStatusExtractFailed || s == RawDocumentsStatusInsertFailed
}

// IsRawDocumentInsertingStatus 是否处于审核入库进行中（含重启后遗留的展示状态）
func IsRawDocumentInsertingStatus(status string) bool {
	return strings.Contains(strings.TrimSpace(status), "正在入库")
}

// IsRawDocumentExtractingStatus 是否处于识别/提取进行中（含重启后遗留的展示状态）
func IsRawDocumentExtractingStatus(status string) bool {
	s := strings.TrimSpace(status)
	if s == "" {
		return false
	}
	if s == RawDocumentsStatusExtracting || s == RawDocumentsStatusUploadedExtracting {
		return true
	}
	return strings.Contains(s, "正在提取") || strings.Contains(s, "正在转文字")
}

// RawDocumentStatusAfterCancelExtract 中断识别后应恢复的文档状态
func RawDocumentStatusAfterCancelExtract(isToMd int64) string {
	if isToMd == 1 {
		return RawDocumentsStatusExtractedNotInDB
	}
	return RawDocumentsStatusUploaded
}

// ResolveRawDocumentListStatus 与列表接口一致：优先保证已审核文档展示为「已经入库」
func ResolveRawDocumentListStatus(status string, isAudit, isToMd int64) string {
	s := strings.TrimSpace(status)
	// 历史数据中可能出现 is_audit=1 但 status 仍是「已提取文字未审核入库」；
	// 这类场景统一按「已经入库」展示，避免前端出现审核后仍未入库的错觉。
	if isAudit == 1 && !strings.Contains(s, "正在出库") {
		return RawDocumentsStatusInserted
	}
	if s != "" {
		return s
	}
	switch {
	case isAudit == 1:
		return RawDocumentsStatusInserted
	case isToMd == 1:
		return RawDocumentsStatusExtractedNotInDB
	default:
		return RawDocumentsStatusUploadedExtracting
	}
}

// IsRawDocumentExtractingBusy 是否应视为「识别进行中」（含 DB status 为空但列表会显示正在提取的遗留数据）
func IsRawDocumentExtractingBusy(status string, isToMd int64) bool {
	if isToMd == 1 {
		return false
	}
	if IsRawDocumentExtractingStatus(status) {
		return true
	}
	if IsRawDocumentFailedStatus(status) {
		return true
	}
	// 与 listrawdocumentslogic 推断一致：status 为空且未转 MD → 界面为「已上传，正在提取文字...」
	return strings.TrimSpace(status) == ""
}
