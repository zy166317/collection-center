package dao

import (
	"collection-center/internal/logger"
	"collection-center/service/db"
	"time"
)

// 项目表
type Project struct {
	Id                 int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	MerchantUid        int64     `json:"merchant_uid" xorm:"not null  comment('商家uid') bigint 'merchant_uid'"`
	ProjectUid         int64     `json:"project_uid" xorm:"unique not null  comment('项目uid') bigint 'project_uid'"` //雪花生成uid,用于跟收款信息表关联
	Name               string    `json:"name" xorm:"not null  comment('项目名称') varchar(255) 'name'"`
	Domain             string    `json:"domain" xorm:"not null default '' comment('项目域名') varchar(255) 'domain'"`
	NotifyUrl          string    `json:"notify_url" xorm:"not null  comment('回调地址') varchar(255) 'notify_url'"`
	ProjectAuditStatus string    `json:"project_audit_status" xorm:"not null default '' comment('审核状态') varchar(64) 'project_audit_status'"`
	ProjectStatus      string    `json:"project_status" xorm:"not null  comment('项目状态') varchar(64) 'project_status'"`
	CreatedAt          time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt          time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}

func (pj *Project) TableName() string {
	return "project"
}

// 定义商家状态
const (
	ProjectStatusNormal = "NORMAL"
	ProjectStatusFreeze = "FREEZE"
)

// 定义商家审核状态
const (
	ProjectAuditStatusPending = "PENDING"
	ProjectAuditStatusPass    = "PASS"
	ProjectAuditStatusReject  = "REJECT"
)

func CreateProject(project *Project, collectRecords []*Collect) (*Project, []*Collect, error) {
	session := db.Client().NewSession()
	_, err := session.Insert(project)
	if err != nil {
		logger.Error("create project error:", err)
		session.Rollback()
		return nil, nil, err
	}
	for _, collect := range collectRecords {
		_, err := session.Insert(collect)
		if err != nil {
			logger.Error("create collect error:", err)
			session.Rollback()
			return nil, nil, err
		}
	}
	session.Commit()
	return project, collectRecords, nil
}

func GetProjectByProjectUid(projectUid int64) (*Project, error) {
	project := &Project{}
	get, err := db.Client().Where("project_uid = ?", projectUid).Get(project)
	if err != nil || !get {
		logger.Error("get project error:", err)
		return nil, err
	}
	return project, nil
}

func UpdateProjectInfo(merchantUid, projectUid int64, domain, notifyUrl string) (int64, error) {
	rows, err := db.Client().Where("merchant_uid = ? and project_uid = ?", merchantUid, projectUid).Update(&Project{Domain: domain, NotifyUrl: notifyUrl})
	if err != nil || rows == 0 {
		logger.Error("update project error:", err)
		return 0, err
	}
	return rows, nil
}

func FreezeProject(merchantUid, projectUid int64) (int64, error) {
	rows, err := db.Client().Where("merchant_uid = ? and project_uid = ?", merchantUid, projectUid).Update(&Project{ProjectStatus: ProjectStatusFreeze})
	if err != nil || rows == 0 {
		logger.Error("freeze project error:", err)
		return rows, err
	}
	return rows, nil
}
