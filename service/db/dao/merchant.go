package dao

import (
	"collection-center/internal/logger"
	"collection-center/service/db"
	"time"
)

// 商家
type Merchant struct {
	Id                  int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	Email               string    `json:"email" xorm:"unique not null  comment('邮箱') varchar(255) 'email'"`
	MerchantUid         int64     `json:"merchant_uid" xorm:"unique  comment('商家uid') bigint 'merchant_uid'"` //雪花生成id,用于跟项目表关联
	Password            string    `json:"password" xorm:"not null  comment('密码') varchar(255) 'password'"`
	MerchantAuditStatus string    `json:"merchant_audit_status" xorm:"not null  comment('审核状态') varchar(64) 'merchant_audit_status'"`
	MerchantStatus      string    `json:"merchant_status" xorm:"not null  comment('商家状态') varchar(64) 'merchant_status'"`
	CreatedAt           time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt           time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}

func (m *Merchant) TableName() string {
	return "merchant"
}

// 定义商家状态
const (
	MerchantStatusNormal = "NORMAL"
	MerchantStatusFreeze = "FREEZE"
)

// 定义商家审核状态
const (
	MerchantAuditStatusPending = "PENDING"
	MerchantAuditStatusPass    = "PASS"
	MerchantAuditStatusReject  = "REJECT"
)

// CreateMerchant 创建商家
func CreateMerchant(merchant *Merchant) (int64, error) {
	merchantId, err := db.Client().Insert(merchant)
	if err != nil {
		logger.Error("insert merchant error:", err)
		return 0, err
	}
	return merchantId, err
}

// IsExistEmail 判断邮箱是否存在
func IsExistEmail(email string) (bool, error) {
	merchant := &Merchant{}
	has, err := db.Client().Where("email = ?", email).Get(merchant)
	if err != nil {
		logger.Error("get merchant error:", err)
		return false, err
	}
	return has, err
}

// GetMerchantByEmail 验证商家账号密码是否匹配
func GetMerchantByEmail(email string) (*Merchant, error) {
	merchant := &Merchant{}
	has, err := db.Client().Where("email = ?", email).Get(merchant)
	if err != nil && !has {
		logger.Error("get merchant error:", err)
		return nil, err
	}
	return merchant, err
}
