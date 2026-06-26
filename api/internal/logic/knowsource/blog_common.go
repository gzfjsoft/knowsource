package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"
)

func normalizeBlogAlias(alias string) string {
	alias = strings.TrimSpace(alias)
	alias = strings.Trim(alias, "/")
	return alias
}

func mapBlogRecordToItem(row *model.BlogRecord, withContent bool) types.BlogItem {
	if row == nil {
		return types.BlogItem{}
	}
	item := types.BlogItem{
		Id:          row.Id,
		Title:       strings.TrimSpace(row.Title),
		Alias:       strings.TrimSpace(row.Alias.String),
		Summary:     strings.TrimSpace(row.Summary),
		Tags:        strings.TrimSpace(row.Tags),
		Categories:  strings.TrimSpace(row.Categories),
		Authors:     strings.TrimSpace(row.Authors),
		Banner:      strings.TrimSpace(row.Banner),
		IsPublished: row.IsPublished,
		CreatedAt:   row.CreatedAt.Unix(),
		UpdatedAt:   row.UpdatedAt.Unix(),
	}
	if withContent {
		item.Content = row.Content
	}
	return item
}

func ensureBlogAdminPermission(ctx context.Context) *types.Response {
	clientId, _ := ctx.Value("clientId").(string)
	if !strings.EqualFold(strings.TrimSpace(clientId), consts.ONLY_ADMIN) || !utils.IsSuperAdminRoleFromContext(ctx) {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "权限不足，仅 admin 租户 superadmin 可操作",
		}
	}
	return nil
}
