// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"strings"
	"unicode/utf8"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	keywordMergeWindow = 2000
	snippetMaxLen      = 2000
	snippetBackwardLen = 400
	snippetForwardLen  = 1600
)

type SearchRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 全文检索原始文档
func NewSearchRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchRawDocumentsLogic {
	return &SearchRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchRawDocumentsLogic) SearchRawDocuments(req *types.SearchRawDocumentsRequest) (resp *types.SearchRawDocumentsResp, err error) {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return &types.SearchRawDocumentsResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "关键字不能为空",
			},
		}, nil
	}

	// pagination
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.SearchRawDocumentsResp{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	rows, err := l.svcCtx.RawDocumentsModel.SearchByKeywordAll(l.ctx, clientId, keyword, req.DocumentCode, req.Tag, req.IsAudit)
	if err != nil {
		l.Logger.Errorf("全文检索原始文档失败: %v, keyword=%s", err, keyword)
		return &types.SearchRawDocumentsResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	expandedRows := expandKeywordMatches(rows, keyword)
	total := int64(len(expandedRows))
	start := int(offset)
	if start > len(expandedRows) {
		start = len(expandedRows)
	}
	end := start + int(pageSize)
	if end > len(expandedRows) {
		end = len(expandedRows)
	}
	pagedRows := expandedRows[start:end]

	list := make([]types.SearchRawDocumentsItem, 0, len(pagedRows))
	for _, r := range pagedRows {
		list = append(list, types.SearchRawDocumentsItem{
			Id:           r.Id,
			DocumentCode: r.DocumentCode,
			FileName:     r.FileName,
			Tag:          r.Tag,
			Snippet:      r.Snippet,
			CreatedAt:    r.CreatedAt.Unix(),
			UpdatedAt:    r.UpdatedAt.Unix(),
			IsAudit:      r.IsAudit,
			IsToMd:       r.IsToMd,
			IsToAi:       r.IsToAi,
			Status:       r.Status,
			StatusMsg:    strings.TrimSpace(r.StatusMsg),
		})
	}

	return &types.SearchRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: types.SearchRawDocumentsData{
			List:  list,
			Total: total,
		},
	}, nil
}

func expandKeywordMatches(rows []*model.RawDocumentsSearchResult, keyword string) []*model.RawDocumentsSearchResult {
	kw := strings.TrimSpace(keyword)
	if kw == "" {
		return rows
	}
	kwLen := utf8.RuneCountInString(kw)

	expanded := make([]*model.RawDocumentsSearchResult, 0, len(rows))
	for _, r := range rows {
		contentMatches := keywordPositions(r.Content, kw)
		if len(contentMatches) > 0 {
			mergedRanges := mergeMatchPositions(contentMatches, kwLen, keywordMergeWindow)
			for _, rg := range mergedRanges {
				item := *r
				item.Snippet = buildRangeSnippet(r.Content, rg[0], rg[1], snippetMaxLen)
				expanded = append(expanded, &item)
			}
			continue
		}

		nameMatches := keywordPositions(r.FileName, kw)
		if len(nameMatches) > 0 {
			for range nameMatches {
				item := *r
				item.Snippet = buildLeadingSnippet(r.Content, snippetMaxLen)
				if item.Snippet == "" {
					item.Snippet = r.FileName
				}
				expanded = append(expanded, &item)
			}
			continue
		}

		if r.Snippet == "" {
			r.Snippet = buildLeadingSnippet(r.Content, snippetMaxLen)
		}
		expanded = append(expanded, r)
	}

	return expanded
}

func mergeMatchPositions(positions []int, keywordLen int, window int) [][2]int {
	if len(positions) == 0 {
		return nil
	}
	if window < 0 {
		window = 0
	}
	if keywordLen <= 0 {
		keywordLen = 1
	}

	ranges := make([][2]int, 0, len(positions))
	currentStart := positions[0]
	currentEnd := positions[0] + keywordLen

	for i := 1; i < len(positions); i++ {
		pos := positions[i]
		if pos-currentEnd <= window {
			end := pos + keywordLen
			if end > currentEnd {
				currentEnd = end
			}
			continue
		}

		ranges = append(ranges, [2]int{currentStart, currentEnd})
		currentStart = pos
		currentEnd = pos + keywordLen
	}

	ranges = append(ranges, [2]int{currentStart, currentEnd})
	return ranges
}

func keywordPositions(text, keyword string) []int {
	if text == "" || keyword == "" {
		return nil
	}
	textRunes := []rune(strings.ToLower(text))
	keywordRunes := []rune(strings.ToLower(keyword))
	if len(keywordRunes) == 0 || len(textRunes) < len(keywordRunes) {
		return nil
	}
	positions := make([]int, 0, 4)
	for start := 0; start <= len(textRunes)-len(keywordRunes); {
		idx := indexRunes(textRunes[start:], keywordRunes)
		if idx < 0 {
			break
		}
		pos := start + idx
		positions = append(positions, pos)
		start = pos + len(keywordRunes)
	}
	return positions
}

func buildSnippet(text string, matchPos int, keywordLen int, maxLen int) string {
	if text == "" {
		return ""
	}
	if maxLen <= 0 {
		maxLen = snippetMaxLen
	}
	runes := []rune(text)
	start, end := windowBounds(len(runes), matchPos, matchPos+keywordLen, maxLen)
	return string(runes[start:end])
}

func buildRangeSnippet(text string, matchStart int, matchEnd int, maxLen int) string {
	if text == "" {
		return ""
	}
	if maxLen <= 0 {
		maxLen = snippetMaxLen
	}
	if matchStart < 0 {
		matchStart = 0
	}
	if matchEnd < matchStart {
		matchEnd = matchStart
	}
	runes := []rune(text)
	if matchEnd > len(runes) {
		matchEnd = len(runes)
	}
	start, end := windowBounds(len(runes), matchStart, matchEnd, maxLen)
	return string(runes[start:end])
}

func buildLeadingSnippet(text string, maxLen int) string {
	if text == "" {
		return ""
	}
	if maxLen <= 0 {
		maxLen = snippetMaxLen
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen])
}

func indexRunes(text []rune, keyword []rune) int {
	if len(keyword) == 0 || len(text) < len(keyword) {
		return -1
	}
	last := len(text) - len(keyword)
	for i := 0; i <= last; i++ {
		match := true
		for j := 0; j < len(keyword); j++ {
			if text[i+j] != keyword[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func windowBounds(textLen int, focusStart int, focusEnd int, maxLen int) (int, int) {
	if textLen <= 0 {
		return 0, 0
	}
	if maxLen <= 0 {
		maxLen = snippetMaxLen
	}
	if textLen <= maxLen {
		return 0, textLen
	}
	if focusStart < 0 {
		focusStart = 0
	}
	if focusEnd < focusStart {
		focusEnd = focusStart
	}
	if focusEnd > textLen {
		focusEnd = textLen
	}

	start := focusStart - snippetBackwardLen
	if start < 0 {
		start = 0
	}
	end := focusStart + snippetForwardLen
	if end > textLen {
		end = textLen
	}

	// 若命中区间超过当前窗口，优先保证完整命中可见，再回退到最长 maxLen。
	if end < focusEnd {
		end = focusEnd
	}
	if end-start > maxLen {
		start = end - maxLen
		if start < 0 {
			start = 0
		}
	}
	return start, end
}
