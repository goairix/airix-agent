package context

import "context"

const (
	AdminIdKey     = "X-Admin-Id"
	WorkspaceIdKey = "X-Workspace-Id"
)

type BizContext struct {
	context.Context
	data BizData
}

type BizData struct {
	AdminID     string `json:"admin_id"`
	WorkspaceID string `json:"workspace_id"`
}

func NewBizContext(ctx context.Context) *BizContext {
	if ctx == nil {
		ctx = context.Background()
	}

	var adminId, workspaceId string

	adminIdCtx := ctx.Value(AdminIdKey)
	workspaceIdCtx := ctx.Value(WorkspaceIdKey)

	if adminIdCtx != nil {
		adminId = adminIdCtx.(string)
	}
	if workspaceIdCtx != nil {
		workspaceId = workspaceIdCtx.(string)
	}

	return &BizContext{
		Context: ctx,
		data: BizData{
			AdminID:     adminId,
			WorkspaceID: workspaceId,
		},
	}
}

func (c *BizContext) Data() BizData {
	return c.data
}
