package svc

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"knowsource/api/internal"
	"knowsource/api/internal/config"
	"knowsource/api/internal/middleware"
	"knowsource/api/internal/utils"
	"knowsource/common/jwtx"
	"knowsource/model"

	"github.com/qdrant/go-client/qdrant"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                  config.Config
	EmailConfig             config.EmailConfig
	Mysql                   sqlx.SqlConn
	Auth                    rest.Middleware
	AuthInterceptor         rest.Middleware
	RedisClient             *redis.Redis
	JWTKeyManager           *jwtx.KeyManager
	W                       *internal.MyWords
	UsersModel              model.UsersModel
	ServerModel             model.ServersModel
	InstanceModel           model.InstancesModel
	RechargeOrderModel      model.RechargeOrdersModel
	TransactionRecordsModel model.TransactionRecordsModel
	VerificationCodesModel  model.VerificationCodesModel
	OrderRecordsModel       model.OrderRecordsModel
	OrganizationModel       model.OrganizationsModel
	BalancesModel           model.BalancesModel
	OrgsUsersModel          model.OrgsUsersModel
	InvitationModel         model.InvitationModel
	ApplyJoinModel          model.ApplyJoinModel
	ServerTagsModel         model.ServerTagsModel
	TagsModel               model.TagsModel
	RolesModel              model.RolesModel
	PermissionsModel        model.PermissionsModel
	RolePermissionsModel    model.RolePermissionsModel
	UserRolesModel          model.UserRolesModel
	DailyUsageModel         model.DailyUsageModel
	MinuteUsageModel        model.MinuteUsageModel
	HourlyUsageModel        model.HourlyUsageModel
	ResourcesModel          model.ResourcesModel
	RegionsModel            model.RegionsModel
	RunningResourcesModel   model.RunningResourcesModel
	WebConfigModel          model.WebConfigModel
	PlayLogModel            model.PlayLogModel
	SongPm5Model            model.SongPm5Model
	FileLikeModel           model.FileLikeModel
	UserBlackListModel      model.UserBlackListModel
	UserAcquireUserModel    model.UserAcquireUserModel
	UserUploadPhotoModel    model.UserUploadPhotoModel
	UsersLoginLogModel      model.UsersLoginLogModel
	UserAuthIdcardModel     model.UserAuthIdcardModel
	// Forum models

	// Knowdata Models
	// UserModel                     knowdata.UserModel
	// RoleModel                     knowdata.RoleModel
	// KnowdataOrganizationModel      knowdata.OrganizationModel
	// CompanyModel                  knowdata.CompanyModel
	// UserAuditLogModel             knowdata.UserAuditLogModel
	// RolePermissionModel           knowdata.RolePermissionModel
	// UserRoleModel                 knowdata.UserRoleModel
	// PermissionModel               knowdata.PermissionModel
	// AnnouncementModel             knowdata.AnnouncementModel
	// UserPointsLogModel            knowdata.UserPointsLogModel
	// BusinessTypeModel             knowdata.BusinessTypeModel
	// AppBannerModel                knowdata.AppBannerModel
	// KnowledgeFeedbackModel        knowdata.KnowledgeFeedbackModel
	// FeedbackAuditLogModel         knowdata.FeedbackAuditLogModel
	// OfflinePackageModel           knowdata.OfflinePackageModel
	// KnowledgeDataFileModel        knowdata.KnowledgeDataFileModel
	// VideoModel                    knowdata.VideoModel
	// VideoChapterModel             knowdata.VideoChapterModel
	// KnowledgeAuditLogModel        knowdata.KnowledgeAuditLogModel
	// FavoriteModel                 knowdata.FavoriteModel
	// LikeModel                     knowdata.LikeModel
	// TechnicalExchangeModel        knowdata.TechnicalExchangeModel
	// TechnicalExchangeAnswersModel knowdata.TechnicalExchangeAnswersModel
	AiConfigModel model.AiConfigModel
	// UserMessageModel          knowdata.UserMessageModel
	// ApkVersionModel           knowdata.ApkVersionModel
	// AsyncTaskModel            knowdata.AsyncTaskModel
	// KnowledgeDetailModel      knowdata.KnowledgeDetailModel
	// KnowledgeDetailDraftModel knowdata.KnowledgeDetailDraftModel
	// SysMenuModel              knowdata.SysMenuModel
	// UserPointsRuleModel       knowdata.UserPointsRuleModel
	// UserPointsRankingModel    knowdata.UserPointsRankingModel
	// ValueAddedServicesModel   knowdata.ValueAddedServicesModel
	// ExchangeGiftModel         knowdata.ExchangeGiftModel
	// MessageModel              knowdata.MessageModel
	// ConversationModel         knowdata.ConversationModel
	// ExpertServicePeriodModel  knowdata.ExpertServicePeriodModel
	// SpecialistModel           knowdata.SpecialistModel
	// UserSpecialistRatingModel knowdata.UserSpecialistRatingModel

	// AI Models
	AiSessionsModel  model.AiSessionsModel
	AiMessagesModel  model.AiMessagesModel
	AiCallStatsModel model.AiCallStatsModel
	BlogModel        model.BlogModel

	// Document Models
	DocumentTypeModel       model.DocumentTypeModel
	RawDocumentsModel       model.RawDocumentsModel
	RawDocumentQaPairsModel model.RawDocumentQaPairsModel

	// knowsource Models
	FrEmpModel              model.FrEmpModel
	FrDeptModel             model.FrDeptModel
	EmpPasswordModel        model.EmpPasswordModel
	EmpDocumentTypeModel    model.EmpDocumentTypeModel
	DeptDocumentTypeModel   model.DeptDocumentTypeModel
	DifyOptionsModel        model.DifyOptionsModel
	FrRolesModel            model.FrRolesModel
	FrPermissionsModel      model.FrPermissionsModel
	FrUserRolesModel        model.FrUserRolesModel
	FrRolesPermissionsModel model.FrRolesPermissionsModel
	FrTagsModel             model.TagsModel
	ClientModel             model.ClientModel

	QdrantClient *qdrant.Client
	QdrantTools  *utils.QdrantTools
}

func NewServiceContext(c config.Config, e config.EmailConfig) (*ServiceContext, error) {
	if c.MySQL.DataSource != "" {
		db, err := sql.Open("mysql", c.MySQL.DataSource)
		if err != nil {
			return nil, fmt.Errorf("mysql open: %w", err)
		}
		if err := db.PingContext(context.Background()); err != nil {
			db.Close()
			return nil, fmt.Errorf("mysql ping: %w", err)
		}
		db.Close()
	}
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	if len(c.CacheRedis) == 0 {
		return nil, fmt.Errorf("CacheRedis 未配置")
	}
	rConf := redis.RedisConf{
		Host: c.CacheRedis[0].Host,
		Type: c.CacheRedis[0].Type,
		Pass: c.CacheRedis[0].Pass,
	}
	redisClient, err := redis.NewRedis(rConf)
	if err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}
	if !redisClient.PingCtx(context.Background()) {
		return nil, fmt.Errorf("redis ping 失败")
	}

	// 初始化JWT密钥管理器
	jwtKeyManager := jwtx.NewKeyManager(redisClient)
	jwtx.SetKeyManager(jwtKeyManager)

	// 初始化 Qdrant 客户端
	var qdrantClient *qdrant.Client
	if c.Qdrant.Host != "" && c.Qdrant.Port > 0 {
		client, err := qdrant.NewClient(&qdrant.Config{
			Host: c.Qdrant.Host,
			Port: c.Qdrant.Port,
		})
		if err != nil {
			// 如果连接失败，记录错误但不中断服务启动
			logx.Errorf("初始化 Qdrant 客户端失败: %v", err)
		} else {
			qdrantClient = client
		}
	}

	var qdrantTools *utils.QdrantTools
	if qdrantClient != nil {
		qdrantTools = utils.NewQdrantToolsWithClient(qdrantClient)
	}

	return &ServiceContext{
		Config:                  c,
		EmailConfig:             e,
		Mysql:                   conn,
		AuthInterceptor:         AuthInterceptor(c.Auth.AccessSecret),
		Auth:                    middleware.NewAuthMiddleware(c, conn, redisClient).Handle,
		RedisClient:             redisClient,
		JWTKeyManager:           jwtKeyManager,
		W:                       internal.NewMyWords(),
		UsersModel:              model.NewUsersModel(conn),
		ServerModel:             model.NewServersModel(conn),
		InstanceModel:           model.NewInstancesModel(conn),
		RechargeOrderModel:      model.NewRechargeOrdersModel(conn),
		TransactionRecordsModel: model.NewTransactionRecordsModel(conn),
		VerificationCodesModel:  model.NewVerificationCodesModel(conn),
		OrderRecordsModel:       model.NewOrderRecordsModel(conn),
		OrganizationModel:       model.NewOrganizationsModel(conn),
		BalancesModel:           model.NewBalancesModel(conn),
		OrgsUsersModel:          model.NewOrgsUsersModel(conn),
		InvitationModel:         model.NewInvitationModel(conn),
		ApplyJoinModel:          model.NewApplyJoinModel(conn),
		ServerTagsModel:         model.NewServerTagsModel(conn),
		TagsModel:               model.NewTagsModel(conn),
		RolesModel:              model.NewRolesModel(conn),
		PermissionsModel:        model.NewPermissionsModel(conn),
		RolePermissionsModel:    model.NewRolePermissionsModel(conn),
		UserRolesModel:          model.NewUserRolesModel(conn),
		DailyUsageModel:         model.NewDailyUsageModel(conn),
		MinuteUsageModel:        model.NewMinuteUsageModel(conn),
		HourlyUsageModel:        model.NewHourlyUsageModel(conn),
		ResourcesModel:          model.NewResourcesModel(conn),
		RegionsModel:            model.NewRegionsModel(conn),
		RunningResourcesModel:   model.NewRunningResourcesModel(conn),
		WebConfigModel:          model.NewWebConfigModel(conn),
		PlayLogModel:            model.NewPlayLogModel(conn),
		SongPm5Model:            model.NewSongPm5Model(conn),
		FileLikeModel:           model.NewFileLikeModel(conn),
		UserBlackListModel:      model.NewUserBlackListModel(conn),
		UserAcquireUserModel:    model.NewUserAcquireUserModel(conn),
		UserUploadPhotoModel:    model.NewUserUploadPhotoModel(conn),
		UsersLoginLogModel:      model.NewUsersLoginLogModel(conn),
		UserAuthIdcardModel:     model.NewUserAuthIdcardModel(conn),

		// UserModel:                     knowdata.NewUserModel(conn),
		// RoleModel:                     knowdata.NewRoleModel(conn),
		// KnowdataOrganizationModel:      knowdata.NewOrganizationModel(conn),
		// CompanyModel:                  knowdata.NewCompanyModel(conn),
		// UserAuditLogModel:             knowdata.NewUserAuditLogModel(conn),
		// RolePermissionModel:           knowdata.NewRolePermissionModel(conn),
		// UserRoleModel:                 knowdata.NewUserRoleModel(conn),
		// PermissionModel:               knowdata.NewPermissionModel(conn),
		// AnnouncementModel:             knowdata.NewAnnouncementModel(conn),
		// UserPointsLogModel:            knowdata.NewUserPointsLogModel(conn),
		// BusinessTypeModel:             knowdata.NewBusinessTypeModel(conn),
		// AppBannerModel:                knowdata.NewAppBannerModel(conn),
		// KnowledgeFeedbackModel:        knowdata.NewKnowledgeFeedbackModel(conn),
		// FeedbackAuditLogModel:         knowdata.NewFeedbackAuditLogModel(conn),
		// OfflinePackageModel:           knowdata.NewOfflinePackageModel(conn),
		// KnowledgeDataFileModel:        knowdata.NewKnowledgeDataFileModel(conn),
		// VideoModel:                    knowdata.NewVideoModel(conn),
		// VideoChapterModel:             knowdata.NewVideoChapterModel(conn),
		// KnowledgeAuditLogModel:        knowdata.NewKnowledgeAuditLogModel(conn),
		// FavoriteModel:                 knowdata.NewFavoriteModel(conn),
		// LikeModel:                     knowdata.NewLikeModel(conn),
		// TechnicalExchangeModel:        knowdata.NewTechnicalExchangeModel(conn),
		// TechnicalExchangeAnswersModel: knowdata.NewTechnicalExchangeAnswersModel(conn),
		AiConfigModel: model.NewAiConfigModel(conn),
		// UserMessageModel:              knowdata.NewUserMessageModel(conn),
		// ApkVersionModel:               knowdata.NewApkVersionModel(conn),
		// AsyncTaskModel:                knowdata.NewAsyncTaskModel(conn),
		// KnowledgeDetailDraftModel:     knowdata.NewKnowledgeDetailDraftModel(conn),
		// KnowledgeDetailModel:          knowdata.NewKnowledgeDetailModel(conn),
		// SysMenuModel:                  knowdata.NewSysMenuModel(conn),
		// UserPointsRuleModel:           knowdata.NewUserPointsRuleModel(conn),
		// UserPointsRankingModel:        knowdata.NewUserPointsRankingModel(conn),
		// ValueAddedServicesModel:       knowdata.NewValueAddedServicesModel(conn),
		// ExchangeGiftModel:             knowdata.NewExchangeGiftModel(conn),
		// MessageModel:                  knowdata.NewMessageModel(conn),
		// ConversationModel:             knowdata.NewConversationModel(conn),
		// ExpertServicePeriodModel:      knowdata.NewExpertServicePeriodModel(conn),
		// SpecialistModel:               knowdata.NewSpecialistModel(conn),
		// UserSpecialistRatingModel:     knowdata.NewUserSpecialistRatingModel(conn),

		AiSessionsModel:  model.NewAiSessionsModel(conn),
		AiMessagesModel:  model.NewAiMessagesModel(conn),
		AiCallStatsModel: model.NewAiCallStatsModel(conn),
		BlogModel:        model.NewBlogModel(conn),

		DocumentTypeModel:       model.NewDocumentTypeModel(conn),
		RawDocumentsModel:       model.NewRawDocumentsModel(conn),
		RawDocumentQaPairsModel: model.NewRawDocumentQaPairsModel(conn),

		FrEmpModel:              model.NewFrEmpModel(conn),
		FrDeptModel:             model.NewFrDeptModel(conn),
		EmpPasswordModel:        model.NewEmpPasswordModel(conn),
		EmpDocumentTypeModel:    model.NewEmpDocumentTypeModel(conn),
		DeptDocumentTypeModel:   model.NewDeptDocumentTypeModel(conn),
		DifyOptionsModel:        model.NewDifyOptionsModel(conn),
		FrRolesModel:            model.NewFrRolesModel(conn),
		FrPermissionsModel:      model.NewFrPermissionsModel(conn),
		FrUserRolesModel:        model.NewFrUserRolesModel(conn),
		FrRolesPermissionsModel: model.NewFrRolesPermissionsModel(conn),
		FrTagsModel:             model.NewTagsModel(conn),
		ClientModel:             model.NewClientModel(conn),
		QdrantClient:            qdrantClient,
		QdrantTools:             qdrantTools,
	}, nil
}

func AuthInterceptor(accessSecret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add your authentication logic here
			// For example, you can check for tokens, validate user sessions, etc.
			// If authentication fails, you can return an error response

			// For now, we'll just call the next handler
			next(w, r)
		}
	}
}
