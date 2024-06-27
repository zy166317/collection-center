package dao

import (
	"collection-center/internal/logger"
	"collection-center/service/db"
	"time"
)

// 收款信息表
type Collect struct {
	Id              int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	MerchantUid     int64     `json:"merchant_uid" xorm:"not null  comment('商家uid') bigint 'merchant_uid'"` //商家uid
	ProjectUid      int64     `json:"project_uid" xorm:"not null  comment('项目id') bigint 'project_uid'"`    //项目表uid
	CollectUid      int64     `json:"collect_uid" xorm:"unique not null  comment('收款信息uid') bigint 'collect_uid'"`
	Chain           string    `json:"chain" xorm:"not null  comment('链名') varchar(64) 'chain'"` //链名
	TokenSymbol     string    `json:"token_symbol" xorm:"not null comment('代币简写') varchar(64) 'token_symbol'"`
	ContractAddress string    `json:"contract_address" xorm:"not null  comment('合约地址') varchar(255) 'contract_address'"` //代币地址
	Decimals        int       `json:"decimals" xorm:"not null  comment('精度') int 'decimals'"`
	LogoUrl         string    `json:"logo_url" xorm:"not null  comment('logo') varchar(255) 'logo_url'"`
	RpcUrl          string    `json:"rpc_url" xorm:"not null  comment('rpc地址') varchar(255) 'rpc_url'"`
	Rate            int       `json:"rate" xorm:"not null  comment('费率') int 'rate'"`
	CollectAddress  string    `json:"collect_address" xorm:"not null  comment('收款地址') varchar(255) 'collect_address'"`
	CreatedAt       time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}

func (c *Collect) TableName() string {
	return "collect"
}

// UpdateCollectsByMerchantUidAndProjectUid 更新收款信息
func UpdateCollectsByMerchantUidAndProjectUid(merchantUid, projectUid int64, chain, collectAddress string) (int64, error) {
	rows, err := db.Client().Where("merchant_uid = ? and project_uid = ? and chain = ? ", merchantUid, projectUid, chain).Update(&Collect{CollectAddress: collectAddress})
	if err != nil {
		logger.Error("UpdateCollectsByMerchantUidAndProjectUid error:", err)
		return 0, err
	}
	if rows == 0 {
		logger.Error("UpdateCollectsByMerchantUidAndProjectUid error: rows == 0")
		return 0, err
	}
	return rows, nil
}

// UpdateCollectRate 更新collect汇率
func UpdateCollectRate(merchantUid, projectUid, collectUid int64, rate int) (int64, error) {
	rows, err := db.Client().Where("merchant_uid = ? and project_uid = ? and collect_uid = ? ", merchantUid, projectUid, collectUid).Update(&Collect{Rate: rate})
	if err != nil {
		logger.Error("UpdateCollectRate error:", err)
		return rows, err
	}
	if rows == 0 {
		logger.Error("UpdateCollectRate error: rows == 0")
		return rows, err
	}
	return rows, nil
}
