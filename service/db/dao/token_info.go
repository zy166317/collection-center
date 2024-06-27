package dao

import (
	"collection-center/internal/logger"
	"collection-center/service/db"
	"time"
)

// token信息表，chain+tokenSymbol做联合索引
type TokenInfo struct {
	Id              int64     `json:"id" xorm:"pk autoincr not null bigint 'id'"`
	Chain           string    `json:"chain" xorm:"not null  comment('链名') varchar(64) 'chain'"`             //链名
	TokenName       string    `json:"token_name" xorm:"not null  comment('代币名称') varchar(64) 'token_name'"` //代币名称
	TokenSymbol     string    `json:"token_symbol" xorm:"not null  comment('代币简写') varchar(64) 'token_symbol'"`
	ContractAddress string    `json:"contract_address" xorm:"not null  comment('代币地址') varchar(255) 'contract_address'"` //代币地址
	Decimals        int       `json:"decimals" xorm:"not null  comment('精度') int 'decimals'"`
	LogoUrl         string    `json:"logo_url" xorm:"not null  comment('logo') varchar(255) 'logo_url'"`
	RpcUrl          string    `json:"rpc_url" xorm:"not null  comment('rpc地址') varchar(255) 'rpc_url'"`
	CreatedAt       time.Time `json:"created_at" xorm:"created" form:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" xorm:"updated" form:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at" xorm:"deleted" form:"deleted_at"`
}

func (t *TokenInfo) TableName() string {
	return "token_info"
}

func GetTokenInfoByChainAndTokenSymbol(chain, tokenSymbol string) (*TokenInfo, error) {
	tokenInfo := &TokenInfo{}
	get, err := db.Client().Where("chain = ? and token_symbol = ?", chain, tokenSymbol).Get(tokenInfo)
	if err != nil || !get {
		return nil, err
	}
	return tokenInfo, nil
}

func CreateTokenInfo(tokenInfo *TokenInfo) error {
	insert, err := db.Client().Insert(tokenInfo)
	if err != nil || insert == 0 {
		logger.Error("create token info error", err)
		return err
	}
	return nil
}
