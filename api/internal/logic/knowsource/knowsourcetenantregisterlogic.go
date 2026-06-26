package knowsource

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/cryptx"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func isMySQLDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	// fallback (in case error is wrapped without preserving type)
	return strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry")
}

type KnowsourceTenantRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourceTenantRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceTenantRegisterLogic {
	return &KnowsourceTenantRegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// KnowsourceTenantRegister 创建租户 + superadmin；status=0 待邮箱验证；clientIP 由 Handler 传入
func (l *KnowsourceTenantRegisterLogic) KnowsourceTenantRegister(req *types.KnowsourceTenantRegisterRequest, clientIP string) (resp *types.KnowsourceTenantRegisterResponse, err error) {
	if req == nil {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "参数不能为空"},
		}, nil
	}
	clientId := strings.TrimSpace(req.ClientId)
	ownerEmail := strings.TrimSpace(strings.ToLower(req.OwnerEmail))
	password := req.Password
	company := strings.TrimSpace(req.CompanyName)
	mobileStr := strings.TrimSpace(req.OwnerMobile)

	if clientId == "" || company == "" || ownerEmail == "" || mobileStr == "" || password == "" {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "clientId、companyName、ownerEmail、ownerMobile、password 不能为空"},
		}, nil
	}
	if len(clientId) <= 6 {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "企业账户名长度必须大于 6 位"},
		}, nil
	}
	if len(password) < 8 {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "密码至少 8 位"},
		}, nil
	}
	// 密码复杂度：大小写字母、数字、符号
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	if !(hasLower && hasUpper && hasDigit && hasSymbol) {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "密码需包含大小写字母、数字和符号"},
		}, nil
	}
	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !reg.MatchString(clientId) {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "企业账户名仅支持字母、数字、下划线和横线"},
		}, nil
	}
	if !strings.Contains(ownerEmail, "@") {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "ownerEmail 格式不正确"},
		}, nil
	}

	if _, e := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, clientId); e == nil {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ConflictCode, Message: "该企业账户名已存在"},
		}, nil
	} else if e != model.ErrNotFound {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()},
		}, nil
	}

	if row, e := l.svcCtx.ClientModel.FindOneByOwnerEmail(l.ctx, ownerEmail); e == nil && row != nil {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ConflictCode, Message: "该负责人邮箱已被其他租户使用"},
		}, nil
	} else if e != nil && e != model.ErrNotFound {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()},
		}, nil
	}

	metaBytes, _ := json.Marshal(map[string]string{
		"name": company,
		"desp": strings.TrimSpace(req.Desp),
	})
	meta := string(metaBytes)

	adm := superadmin.SuperadminEmpCode()
	hash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, password)

	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(s)
		cm := model.NewClientModel(conn)
		em := model.NewFrEmpModel(conn)
		pm := model.NewEmpPasswordModel(conn)
		dm := model.NewFrDeptModel(conn)
		acm := model.NewAiConfigModel(conn)
		dtm := model.NewDocumentTypeModel(conn)

		// 1. 插入租户信息
		if _, e := cm.Insert(ctx, &model.Client{
			ClientId:       clientId,
			ClientJsonInfo: meta,
			RegistrationIp: clientIP,
			Status:         0,
			VerifiedAt:     sql.NullTime{},
			OwnerEmail:     ownerEmail,
		}); e != nil {
			return fmt.Errorf("insert client: %w", e)
		}

		// 2. 创建默认部门 - 办公室
		deptCode := "office"
		if _, e := dm.Insert(ctx, &model.FrDept{
			ClientId:   clientId,
			Gsdm:       clientId,
			DeptCode:   deptCode,
			DeptName:   "办公室",
			ParentCode: "",
			EndMark:    "1",
			GRADE:      1,
			Kind:       "部门",
			B0110:      "",
		}); e != nil {
			return fmt.Errorf("insert department: %w", e)
		}

		// 3. 插入超级管理员，部门编码设置为默认创建的办公室
		if _, e := em.Insert(ctx, &model.FrEmp{
			ClientId:  clientId,
			Frylb:     "在职人员",
			Status:    0,
			FempName:  "超级管理员",
			DeptCode:  deptCode,
			FdeptId:   sql.NullInt64{},
			FempCode:  adm,
			Fposition: "超级管理员",
			Fbranch:   "平台:" + clientId,
			Mobile:    mobileStr,
			Email:     ownerEmail,
		}); e != nil {
			if !isMySQLDuplicateEntry(e) {
				return fmt.Errorf("insert superadmin: %w", e)
			}
		}

		// 4. 插入密码
		if _, e := pm.Insert(ctx, &model.EmpPassword{
			ClientId: clientId,
			EmpCode:  adm,
			Password: hash,
		}); e != nil {
			if !isMySQLDuplicateEntry(e) {
				return fmt.Errorf("insert password: %w", e)
			}
		}

		// 5. 插入默认AI配置
		// 5.1 检索提示词
		if _, e := acm.Insert(ctx, &model.AiConfig{
			ClientId:     clientId,
			DocumentCode: "",
			Name:         "检索提示词",
			Value:        "%s --- 根据以上参考资料，使用原文回答问题 \"%s\" ； 如果以上参考资料中没有问题的答案，就说明找不到这个问题的答案。",
			CreatedBy:    "system",
		}); e != nil {
			return fmt.Errorf("insert ai retrieval prompt: %w", e)
		}

		// 5.2 角色提示词
		rolePrompt := `你是公司的内部智能助手，专注于为员工提供与公司业务相关的专业支持。 请始终遵循以下原则： 1. **身份明确**：你是公司AI助手，始终确保回答符合国有企业合规要求，不得泄露敏感信息或作出越权解释，仅提供信息参考与辅助决策。 2. **内容准确**：优先引用国家/行业标准（如GB/T、LY/T等）、公司内部制度或权威技术资料；若不确定，请明确说明“信息未核实”或“建议咨询相关部门”。 3. **语言规范**：使用正式、简洁、专业的中文，避免口语化、夸张或主观评价。 4. **安全合规**：不处理涉密信息；不生成违反国家法律法规、国企纪律或环保政策的内容；涉及数据时默认脱敏。 5. **实用导向**：回答应聚焦解决实际工作问题，如工艺参数建议、设备故障排查思路、标准条款查询、公文写作辅助等。 当用户提问超出业务范围（如金融投资、个人生活等），请礼貌引导回公司相关业务场景。 现在，请根据用户的具体问题提供专业、可靠、合规的协助。`
		if _, e := acm.Insert(ctx, &model.AiConfig{
			ClientId:     clientId,
			DocumentCode: "",
			Name:         "角色提示词",
			Value:        rolePrompt,
			CreatedBy:    "system",
		}); e != nil {
			return fmt.Errorf("insert ai role prompt: %w", e)
		}

		// 5.3 问候词
		if _, e := acm.Insert(ctx, &model.AiConfig{
			ClientId:     clientId,
			DocumentCode: "",
			Name:         "问候词",
			Value:        "你好，我是AI小助手，请问有什么需要帮助",
			CreatedBy:    "system",
		}); e != nil {
			return fmt.Errorf("insert ai greeting: %w", e)
		}

		// 5.4 问答提取提示词（文档审核入库时用于抽取 Q/A）
		if _, e := acm.Insert(ctx, &model.AiConfig{
			ClientId:     clientId,
			DocumentCode: "",
			Name:         "问答提取提示词",
			Value:        "请阅读以下文本片段，站在用户角度提炼该片段能回答的问题及对应答案。请严格按如下格式输出，至少0组，最多5组：\nQ1: xxx\nA1: xxx\nQ2: xxx\nA2: xxx\n不要输出其他解释。\n\n文本片段：\n%s",
			CreatedBy:    "system",
		}); e != nil {
			return fmt.Errorf("insert ai qa extraction prompt: %w", e)
		}

		// 6. 创建默认企业知识库
		if _, e := dtm.Insert(ctx, &model.DocumentType{
			ClientId:    clientId,
			Code:        "company_knowledge",
			Name:        "企业知识库",
			IsDisabled:  0,
			Description: "企业默认知识库",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}); e != nil {
			return fmt.Errorf("insert default document type: %w", e)
		}

		return nil
	})
	if err != nil {
		logx.Errorf("KnowsourceTenantRegister tx: %v", err)
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "注册失败", Info: err.Error()},
		}, nil
	}

	if e := superadmin.EnsureSuperadminRoleBindings(l.ctx, l.svcCtx, clientId); e != nil {
		logx.Errorf("EnsureSuperadminRoleBindings: %v", e)
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "初始化角色失败", Info: e.Error()},
		}, nil
	}

	code := KnowsourceRandomDigitCode()
	if e := KnowsourceStoreVerificationCode(l.ctx, l.svcCtx, targetKnowsourceTenantVerify, clientId, code, clientIP); e != nil {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "写入验证码失败", Info: e.Error()},
		}, nil
	}

	username := adm
	subject := "【知源智库 AI】验证您的租户邮箱"
	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h2 style="color: #333;">租户邮箱验证</h2>
<p>您的企业账户已创建，请使用以下信息登录（验证邮箱后生效）：</p>
<ul>
<li><strong>企业账户名</strong>：<code>%s</code></li>
<li><strong>用户名</strong>：<code>%s</code></li>
</ul>
<p>您的验证码为：</p>
<div style="background-color: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px;">
<p style="font-size: 24px; font-weight: bold; text-align: center; color: #333;">%s</p>
</div>
<p>验证码 15 分钟内有效，请及时完成验证。</p>
</div>`, clientId, username, code)
	if e := KnowsourceSendSimpleMail(l.svcCtx, ownerEmail, subject, body); e != nil {
		return &types.KnowsourceTenantRegisterResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "租户已创建但验证邮件发送失败，请检查 Mail 配置或联系管理员", Info: e.Error()},
			Data: &types.KnowsourceTenantRegisterData{
				ClientId:   clientId,
				Username:   username,
				NeedVerify: true,
			},
		}, nil
	}

	return &types.KnowsourceTenantRegisterResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success"},
		Data: &types.KnowsourceTenantRegisterData{
			ClientId:   clientId,
			Username:   username,
			NeedVerify: true,
		},
	}, nil
}
